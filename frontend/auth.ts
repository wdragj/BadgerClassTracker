import NextAuth from "next-auth";
import GoogleProvider from "next-auth/providers/google";

export const { handlers, signIn, signOut, auth } = NextAuth({
    providers: [
        GoogleProvider({
            clientId: process.env.GOOGLE_CLIENT_ID!,
            clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
        }),
    ],
    callbacks: {
        async signIn({ user, profile }) {
            try {
                const googleSub = profile?.sub;

                await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/register`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        user: {
                            name: user.name,
                            email: user.email,
                            image: user.image,
                            googleSub,
                        },
                    }),
                });

                // If we successfully stored the user, return true to allow sign-in
                return true;
            } catch (err) {
                // eslint-disable-next-line no-console
                console.error("Failed to store user in DB:", err);

                // Return false to block the sign-in if something goes wrong
                return false;
            }
        },
    },
});
