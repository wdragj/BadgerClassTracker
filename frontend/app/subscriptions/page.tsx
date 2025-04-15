"use client";

import React, { useEffect, useState } from "react";
import { Card, CardBody, Button, Spinner, Alert } from "@heroui/react";
import NotificationsOffIcon from "@mui/icons-material/NotificationsOff";
import useSWR from "swr";
import { useSession } from "next-auth/react";

// Reuse the same interfaces from your home page:
interface Subscription {
    courseId: string;
    courseSubjectCode: string;
    courseName: string;
    credits: number;
    title: string;
}

export default function MySubscriptionsPage() {
    const { data: session, status } = useSession();
    const isAuthenticated = status === "authenticated";

    // Alert states
    const [alertVisible, setAlertVisible] = useState(false);
    const [alertTitle, setAlertTitle] = useState("");
    const [alertDescription, setAlertDescription] = useState("");
    const [alertColor, setAlertColor] = useState<"success" | "danger" | "default" | "primary" | "secondary" | "warning">("success");

    // Fetch the user’s subscriptions
    const {
        data: subsData,
        isLoading,
        mutate,
    } = useSWR<{ subscriptions: Subscription[] }>(
        isAuthenticated && session?.user?.email ? [`subscriptions`, session.user.email] : null,
        () => fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/subscriptions?userEmail=${session?.user?.email}`).then((res) => res.json()),
        { refreshInterval: 60000 }
    );

    // Subscriptions array from the backend
    const subscriptions = subsData?.subscriptions || [];

    // Auto-hide alerts after a short delay
    useEffect(() => {
        if (alertVisible) {
            const timer = setTimeout(() => {
                setAlertVisible(false);
            }, 3000);

            return () => clearTimeout(timer);
        }
    }, [alertVisible]);

    // Unsubscribe logic
    async function handleUnsubscribe(subscription: Subscription) {
        if (!session?.user?.email) {
            console.error("User is not authenticated properly");

            return;
        }

        // Payload for unsubscribe
        const payload = {
            userEmail: session.user.email,
            courseId: subscription.courseId,
            courseSubjectCode: subscription.courseSubjectCode,
        };

        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/unsubscribe`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                console.error("Unsubscription failed:", response.status);
            } else {
                console.log("Unsubscription successful");
                mutate(); // re-fetch updated subscriptions
                setAlertTitle("Unsubscription Successful");
                setAlertDescription(`You have unsubscribed from ${subscription.courseName}.`);
                setAlertColor("success");
                setAlertVisible(true);
            }
        } catch (error) {
            console.error("Error during unsubscription:", error);
        }
    }

    if (!isAuthenticated) {
        return (
            <div className="flex flex-col items-center justify-center h-screen">
                <p className="text-lg font-semibold">Please sign in to view your subscriptions.</p>
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className="flex justify-center items-center h-screen">
                <Spinner label="Loading..." />
            </div>
        );
    }

    return (
        <div className="relative min-h-screen">
            {/* Alert for unsubscribing */}
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

            <div className="py-4 px-3">
                <h1 className="text-2xl font-semibold mb-4 text-center">Subscribed Courses</h1>

                {subscriptions.length === 0 ? (
                    <p>You have no subscriptions.</p>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {subscriptions.map((subscription) => {
                            return (
                                <Card
                                    key={`${subscription.courseId}-${subscription.courseSubjectCode}`}
                                    className="border-b border-gray-200"
                                    radius="none"
                                    shadow="none"
                                >
                                    <CardBody>
                                        {/* First row: courseName + credits */}
                                        <div className="flex justify-between items-center w-full">
                                            <h2 className="text-lg font-semibold">{subscription.courseName}</h2>
                                            <span className="text-sm font-semibold">{subscription.credits} Credit</span>
                                        </div>

                                        {/* Second row: title + Unsubscribe button */}
                                        <div className="flex justify-between items-center w-full">
                                            <p className="text-sm text-gray-600">{subscription.title}</p>
                                            <Button
                                                isIconOnly
                                                color="danger"
                                                radius="full"
                                                size="sm"
                                                variant="flat"
                                                onPress={() => handleUnsubscribe(subscription)}
                                            >
                                                <NotificationsOffIcon fontSize="small" />
                                            </Button>
                                        </div>
                                    </CardBody>
                                </Card>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}
