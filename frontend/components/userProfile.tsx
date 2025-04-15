"use client";

import { Avatar, Popover, PopoverTrigger, PopoverContent, Link } from "@heroui/react";
import { signOut } from "next-auth/react";
import { useState } from "react";

interface Session {
    user?: {
        name?: string | null;
        email?: string | null;
        image?: string | null;
    };
}

interface UserProfileProps {
    session: Session | null;
}

export default function UserProfile({ session }: UserProfileProps) {
    const [isOpen, setIsOpen] = useState(false);

    // If there's no session or user, render nothing
    if (!session || !session.user) return null;

    const avatarSrc = session.user.image || "";

    const handleSignOut = async () => {
        await signOut();
        setIsOpen(false);
        window.location.reload(); // Optionally refresh after sign-out
    };

    // Handles link clicks
    const handleClose = () => setIsOpen(false);

    return (
        <Popover isOpen={isOpen} offset={10} placement="bottom-end" onOpenChange={setIsOpen}>
            <PopoverTrigger>
                <Avatar isBordered as="button" color="danger" size="sm" src={avatarSrc} />
            </PopoverTrigger>
            <PopoverContent className="px-4 py-3 min-w-fit">
                <div className="flex flex-col gap-1">
                    <div className="text-small font-bold">{session.user.name || "User"}</div>
                    <div className="text-xs text-gray-500">{session.user.email}</div>
                    <Link className="text-sm text-gray-600 pt-2 pb-1 inline-flex" href="/" onPress={handleClose}>
                        Home
                    </Link>
                    <Link className="text-sm text-gray-600 pt-2 pb-1 inline-flex" href="/subscriptions" onPress={handleClose}>
                        My Subscriptions
                    </Link>
                    <Link as="button" className="text-sm text-danger pt-2 pb-1 inline-flex" onPress={handleSignOut}>
                        Sign Out
                    </Link>
                </div>
            </PopoverContent>
        </Popover>
    );
}
