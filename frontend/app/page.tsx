"use client";

import React, { useState } from "react";
import { Card, CardBody, Button, Pagination, Spinner, Input } from "@heroui/react";
import NotificationsNoneIcon from "@mui/icons-material/NotificationsNone";
import useSWR from "swr";

import { fetchCourses } from "@/lib/api";

// Course object
interface Course {
    id: string;
    subjectCode: string;
    name: string;
    title: string;
    credits: number;
}

export default function CoursesPage() {
    const [page, setPage] = useState(1);
    const [searchQuery, setSearchQuery] = useState(""); // User input
    const [submittedQuery, setSubmittedQuery] = useState(""); // Actual query for fetching

    const rowsPerPage = 50;

    // Fetch courses using submitted search query
    const { data, isLoading, mutate } = useSWR<{ hits: Course[]; found: number }>(
        ["courses", page, rowsPerPage, submittedQuery],
        () => fetchCourses(page, rowsPerPage, submittedQuery),
        { keepPreviousData: true }
    );

    const paginatedData = data?.hits || [];
    const totalResults = data?.found || 0;
    const pages = Math.ceil(totalResults / rowsPerPage);
    const loadingState = isLoading || !data ? "loading" : "idle";

    // Handle search submission
    const handleSearchSubmit = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") {
            setSubmittedQuery(searchQuery); // Set query for fetching
            setPage(1); // Reset to first page on new search
            mutate(); // Manually trigger fetch
        }
    };

    return (
        <div>
            <div className="sticky top-0 z-10 flex justify-between items-center px-3 w-full bg-background shadow-sm">
                <div className="flex flex-col justify-start text-left">
                    <h1 className="text-lg font-bold mb-1">Courses</h1>
                    <p className="mb-2 text-sm text-gray-600">Total Results: {totalResults}</p>
                </div>

                <div className="w-64">
                    <Input
                        isClearable
                        color="primary"
                        placeholder="Search courses..."
                        value={searchQuery}
                        variant="bordered"
                        onChange={(e) => setSearchQuery(e.target.value)} // Update input
                        onClear={() => setSearchQuery("")} // Clear input
                        onKeyDown={handleSearchSubmit} // Fetch on Enter
                    />
                </div>
            </div>
            {loadingState === "loading" ? (
                <div className="flex justify-center items-center w-full">
                    <Spinner label="Loading..." />
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
                    {paginatedData.map((course: Course) => (
                        <Card key={course.id + course.subjectCode} className="border-b border-gray-200" radius="none" shadow="none">
                            <CardBody>
                                <div className="flex justify-between items-center w-full">
                                    <h2 className="text-lg font-semibold">{course.name}</h2>
                                    <span className="text-sm font-semibold">{course.credits} Credit</span>
                                </div>

                                <div className="flex justify-between items-center w-full">
                                    <p className="text-sm text-gray-600">{course.title}</p>
                                    <Button isIconOnly color="warning" radius="full" size="sm" variant="flat">
                                        <NotificationsNoneIcon fontSize="small" />
                                    </Button>
                                </div>
                            </CardBody>
                        </Card>
                    ))}
                </div>
            )}
            {pages > 0 && (
                <div className="sticky bottom-0 w-full py-1 flex justify-center">
                    <Pagination isCompact showControls showShadow color="danger" page={page} total={pages} onChange={(newPage) => setPage(newPage)} />
                </div>
            )}
        </div>
    );
}
