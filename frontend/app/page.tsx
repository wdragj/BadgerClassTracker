"use client";

import React, { useEffect, useState } from "react";
import { Card, CardBody, Button, Pagination, Spinner, Input, Alert } from "@heroui/react";
import NotificationsNoneIcon from "@mui/icons-material/NotificationsNone";
import NotificationsOffIcon from "@mui/icons-material/NotificationsOff";
import useSWR from "swr";
import { useDisclosure } from "@heroui/react";
import { useSession } from "next-auth/react";

import { fetchCourses } from "@/lib/api";
import SubscribeModal from "@/components/subscribeModal";
import UnsubscribeModal from "@/components/unsubscribeModal";

// Course interface
export interface Course {
    id: string;
    subjectCode: string;
    name: string;
    title: string;
    credits: number;
}

// Subscription interface
interface Subscription {
    courseId: string;
    courseSubjectCode: string;
}

// CoursesResponse interface
interface CoursesResponse {
    term: {
        termCode: string;
        longDescription: string;
    };
    courses: {
        hits: Course[];
        found: number;
    };
}

export default function CoursesPage() {
    const { data: session, status } = useSession();
    const isAuthenticated = status === "authenticated";
    const [page, setPage] = useState(1);
    const [searchQuery, setSearchQuery] = useState(""); // user input
    const [submittedQuery, setSubmittedQuery] = useState(""); // actual query used for fetching
    const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);

    // Alert state for subscription notifications
    const [alertVisible, setAlertVisible] = useState(false);
    const [alertTitle, setAlertTitle] = useState("");
    const [alertDescription, setAlertDescription] = useState("");
    const [alertColor, setAlertColor] = useState<"success" | "danger" | "default" | "primary" | "secondary" | "warning">("success");

    // useDisclosure for subscribe and unsubscribe modals
    const { isOpen: isSubscribeOpen, onOpen: onSubscribeOpen, onClose: onSubscribeClose } = useDisclosure();
    const { isOpen: isUnsubscribeOpen, onOpen: onUnsubscribeOpen, onClose: onUnsubscribeClose } = useDisclosure();

    const rowsPerPage = 50;

    // Fetch courses using SWR
    const { data, isLoading, mutate } = useSWR<{ hits: Course[]; found: number; term: { longDescription: string } }>(
        ["courses", page, rowsPerPage, submittedQuery],
        () => fetchCourses(page, rowsPerPage, submittedQuery),
        { keepPreviousData: true }
    );

    // If authenticated, fetch subscriptions for the current user
    const { data: subsData, mutate: mutateSubscriptions } = useSWR<{ subscriptions: Subscription[] }>(
        isAuthenticated && session?.user?.email ? [`subscriptions`, session.user.email] : null,
        () => fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/subscriptions?userEmail=${session?.user?.email}`).then((res) => res.json()),
        { refreshInterval: 60000 }
    );

    const paginatedData = data?.hits || [];
    const totalResults = data?.found || 0;
    const termLongDescription = data?.term.longDescription || "";

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

    // When the notification icon is clicked for subscribing, open the subscribe modal
    const handleSubscribeClick = (course: Course) => {
        setSelectedCourse(course);
        onSubscribeOpen();
    };

    // When the notification icon is clicked for unsubscribing, open the unsubscribe modal
    const handleUnsubscribeClick = (course: Course) => {
        setSelectedCourse(course);
        onUnsubscribeOpen();
    };

    // Handle subscribe logic
    const handleSubscribe = async (course: Course) => {
        if (!session?.user?.email || !session?.user?.name) {
            // eslint-disable-next-line no-console
            console.error("User is not authenticated properly");

            return;
        }
        const payload = {
            userEmail: session.user.email,
            userFullName: session.user.name,
            courseId: course.id,
            courseName: course.name,
            courseSubjectCode: course.subjectCode,
            // courseStatus is omitted; backend will default it to "open"
        };

        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/subscribe`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                // eslint-disable-next-line no-console
                console.error("Subscription failed:", response.status);
            } else {
                // eslint-disable-next-line no-console
                console.log("Subscription successful");
                // Revalidate subscriptions so UI updates immediately
                mutateSubscriptions();
                // Show success alert
                setAlertTitle("Subscription Successful");
                setAlertDescription(`You have subscribed to ${course.name}.`);
                setAlertColor("success");
                setAlertVisible(true);
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error("Error during subscription:", error);
        }
    };

    // Handle unsubscribe logic
    const handleUnsubscribe = async (course: Course) => {
        if (!session?.user?.email) {
            // eslint-disable-next-line no-console
            console.error("User is not authenticated properly");

            return;
        }
        const payload = {
            userEmail: session.user.email,
            courseId: course.id,
            courseSubjectCode: course.subjectCode,
        };

        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/unsubscribe`, {
                method: "POST", // Or DELETE if your endpoint supports it
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                // eslint-disable-next-line no-console
                console.error("Unsubscription failed:", response.status);
            } else {
                // eslint-disable-next-line no-console
                console.log("Unsubscription successful");
                // Revalidate subscriptions so UI updates immediately
                mutateSubscriptions();
                // Show alert for unsubscription
                setAlertTitle("Unsubscription Successful");
                setAlertDescription(`You have unsubscribed from ${course.name}.`);
                setAlertColor("success");
                setAlertVisible(true);
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error("Error during unsubscription:", error);
        }
    };

    // Check if a course is subscribed
    const isCourseSubscribed = (course: Course): boolean => {
        if (!subsData?.subscriptions) return false;

        return subsData.subscriptions.some((sub) => sub.courseId === course.id && sub.courseSubjectCode === course.subjectCode);
    };

    // When isVisible is true, set a timeout to hide it after 3 seconds
    useEffect(() => {
        if (alertVisible) {
            const timer = setTimeout(() => {
                setAlertVisible(false);
            }, 3000); // 3 seconds

            // Cleanup the timer if the component unmounts or isVisible changes
            return () => clearTimeout(timer);
        }
    }, [alertVisible]);

    return (
        <div className="relative min-h-screen">
            {/* Optional Alert */}
            {alertVisible && (
                <div className="fixed bottom-32 left-0 right-0 z-50 mx-auto max-w-7xl">
                    <Alert
                        color={alertColor}
                        description={alertDescription}
                        isVisible={alertVisible}
                        title={alertTitle}
                        variant="faded"
                        onClose={() => setAlertVisible(false)}
                    />
                </div>
            )}

            {/* Sticky top search bar */}
            <div className="sticky top-16 z-10 flex justify-between items-center px-3 py-1 w-full bg-background shadow-sm">
                <div className="flex flex-col justify-start text-left">
                    <p className="text-sm text-gray-600">{termLongDescription}</p>
                    <p className="text-sm text-gray-600">{totalResults} results</p>
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
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-4">
                    {paginatedData.map((course) => {
                        const subscribed = isCourseSubscribed(course);

                        return (
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
                                            color={isAuthenticated ? (subscribed ? "danger" : "success") : "success"}
                                            radius="full"
                                            size="sm"
                                            variant="flat"
                                            onPress={() => {
                                                if (!isAuthenticated) {
                                                    // Instead of window.alert, we trigger the HeroUI Alert
                                                    setAlertTitle("Sign In Required");
                                                    setAlertDescription("Please sign in to subscribe.");
                                                    setAlertColor("danger");
                                                    setAlertVisible(true);

                                                    return;
                                                }
                                                if (subscribed) {
                                                    handleUnsubscribeClick(course);
                                                } else {
                                                    handleSubscribeClick(course);
                                                }
                                            }}
                                        >
                                            {isAuthenticated && subscribed ? (
                                                <NotificationsOffIcon fontSize="small" />
                                            ) : (
                                                <NotificationsNoneIcon fontSize="small" />
                                            )}
                                        </Button>
                                    </div>
                                </CardBody>
                            </Card>
                        );
                    })}
                </div>
            )}

            {pages > 0 && (
                <div className="sticky bottom-0 z-10 w-full pt-3 pb-6 flex justify-center bg-background">
                    <Pagination isCompact showControls showShadow color="danger" page={page} total={pages} onChange={(newPage) => setPage(newPage)} />
                </div>
            )}

            {/* Render the SubscribeModal if a course is selected and not subscribed */}
            {selectedCourse && !isCourseSubscribed(selectedCourse) && (
                <SubscribeModal
                    course={selectedCourse}
                    isOpen={isSubscribeOpen}
                    onClose={() => {
                        onSubscribeClose();
                        setSelectedCourse(null);
                    }}
                    onSubscribe={handleSubscribe}
                />
            )}

            {/* Render the UnsubscribeModal if a course is selected and is subscribed */}
            {selectedCourse && isCourseSubscribed(selectedCourse) && (
                <UnsubscribeModal
                    course={selectedCourse}
                    isOpen={isUnsubscribeOpen}
                    onClose={() => {
                        onUnsubscribeClose();
                        setSelectedCourse(null);
                    }}
                    onUnsubscribe={handleUnsubscribe}
                />
            )}
        </div>
    );
}
