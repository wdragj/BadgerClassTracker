"use client";

import { Navbar as HeroUINavbar, NavbarContent, NavbarBrand, NavbarItem } from "@heroui/navbar";
import { useSession } from "next-auth/react";
import Link from "next/link";
import SignInButton from "./signInButton";
import UserProfile from "./userProfile";
import { ThemeSwitch } from "@/components/theme-switch";
import { Logo } from "@/components/icons";

export const Navbar = () => {
    const { data: session, status } = useSession();
    const isAuthenticated = status === "authenticated";

    return (
        <HeroUINavbar maxWidth="xl" position="sticky">
            {/* Left Section: Brand */}
            <NavbarContent className="flex-1" justify="start">
                <NavbarBrand as="li" className="gap-3 max-w-fit">
                    <Link className="flex items-center gap-1" href="/">
                        <Logo />
                        <p className="font-bold text-inherit">BCC</p>
                    </Link>
                </NavbarBrand>
            </NavbarContent>

            {/* Center Section: Nav Items (hidden on small screens) */}
            <NavbarContent className="hidden sm:flex flex-1" justify="center">
                <NavbarItem>
                    <Link href="/" className="hover:underline">
                        Home
                    </Link>
                </NavbarItem>
                <NavbarItem>
                    <Link href="/subscriptions" className="hover:underline">
                        My Subscriptions
                    </Link>
                </NavbarItem>
            </NavbarContent>

            {/* Right Section: Theme Switch + Profile/Sign In */}
            <NavbarContent className="flex-1" justify="end">
                <NavbarItem className="hidden sm:flex gap-2">
                    <ThemeSwitch />
                </NavbarItem>
                <NavbarItem>{isAuthenticated ? <UserProfile session={session} /> : <SignInButton />}</NavbarItem>
            </NavbarContent>
        </HeroUINavbar>
    );
};
