"use client";

import React, { useState } from "react";
import { Card, CardBody, Button, Pagination, Spinner, Input } from "@heroui/react";
import NotificationsNoneIcon from "@mui/icons-material/NotificationsNone";
import useSWR from "swr";
import { useDisclosure } from "@heroui/react";
import { fetchCourses } from "@/lib/api";
import SubscribeModal from "@/components/subscribeModal";

// Define your Course interface (adjust as needed)
export interface Course {
    id: string;
    subjectCode: string;
    name: string;
    title: string;
    credits: number;
}

export default function CoursesPage() {
    const [page, setPage] = useState(1);
    const [searchQuery, setSearchQuery] = useState(""); // user input
    const [submittedQuery, setSubmittedQuery] = useState(""); // actual query used for fetching
    const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);

    const { isOpen, onOpen, onClose } = useDisclosure(); // HeroUI hook for modal state

    const rowsPerPage = 50;

    // Fetch courses using SWR
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
            setSubmittedQuery(searchQuery);
            setPage(1);
            mutate();
        }
    };

    // When the notification icon is clicked, open the modal
    const handleSubscribeClick = (course: Course) => {
        setSelectedCourse(course);
        onOpen(); // open modal using useDisclosure
    };

    // Handle subscription logic
    const handleSubscribe = (course: Course) => {
        console.log("Subscribing to course:", course);
        // Insert your subscription logic here (e.g., API call)
        // After subscribing, you might refresh data or show a success message.
    };

    return (
        <div className="relative min-h-screen">
            <div className="sticky top-0 z-10 flex justify-between items-center px-3 pb-1 w-full bg-background shadow-sm">
                <div className="flex flex-col justify-start text-left">
                    <p className="text-sm text-gray-600">Total Results: {totalResults}</p>
                </div>
                <div className="w-64">
                    <Input
                        isClearable
                        color="primary"
                        placeholder="Search courses..."
                        value={searchQuery}
                        variant="bordered"
                        onChange={(e) => setSearchQuery(e.target.value)}
                        onClear={() => setSearchQuery("")}
                        onKeyDown={handleSearchSubmit}
                    />
                </div>
            </div>

            {loadingState === "loading" ? (
                <div className="flex justify-center items-center w-full">
                    <Spinner label="Loading..." />
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {paginatedData.map((course) => (
                        <Card key={course.id + course.subjectCode} className="border-b border-gray-200" radius="none" shadow="none">
                            <CardBody>
                                <div className="flex justify-between items-center w-full">
                                    <h2 className="text-lg font-semibold">{course.name}</h2>
                                    <span className="text-sm font-semibold">{course.credits} Credit</span>
                                </div>
                                <div className="flex justify-between items-center w-full">
                                    <p className="text-sm text-gray-600">{course.title}</p>
                                    <Button
                                        isIconOnly
                                        color="warning"
                                        radius="full"
                                        size="sm"
                                        variant="flat"
                                        onPress={() => handleSubscribeClick(course)}
                                    >
                                        <NotificationsNoneIcon fontSize="small" />
                                    </Button>
                                </div>
                            </CardBody>
                        </Card>
                    ))}
                </div>
            )}

            {pages > 0 && (
                <div className="sticky bottom-0 z-10 w-full py-1 flex justify-center bg-background">
                    <Pagination isCompact showControls showShadow color="danger" page={page} total={pages} onChange={(newPage) => setPage(newPage)} />
                </div>
            )}

            {selectedCourse && (
                <SubscribeModal
                    isOpen={isOpen}
                    course={selectedCourse}
                    onClose={() => {
                        onClose();
                        setSelectedCourse(null);
                    }}
                    onSubscribe={handleSubscribe}
                />
            )}
        </div>
    );
}
