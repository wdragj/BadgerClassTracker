"use client";

import { Button } from "@heroui/react";
import { signIn } from "next-auth/react";

export default function SignInButton() {
    return (
        <Button color="danger" size="md" variant="flat" onPress={() => signIn("google")}>
            Sign In
        </Button>
    );
}
