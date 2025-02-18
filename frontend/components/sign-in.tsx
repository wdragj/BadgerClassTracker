"use client";

import { Button } from "@heroui/react";
import { signIn } from "next-auth/react";

export default function SignIn() {
    return (
        <Button color="primary" size="md" variant="flat" onPress={() => signIn("google")}>
            Sign In
        </Button>
    );
}
