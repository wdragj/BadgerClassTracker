import "@/styles/globals.css";
import { Metadata, Viewport } from "next";
import clsx from "clsx";

import { Providers } from "./providers";

import { siteConfig } from "@/config/site";
import { fontSans } from "@/config/fonts";
import { Navbar } from "@/components/navbar";

export const metadata: Metadata = {
    title: {
        default: siteConfig.name,
        template: `%s - ${siteConfig.name}`,
    },
    description: siteConfig.description,
    icons: {
        icon: "/favicon.ico",
    },
};

export const viewport: Viewport = {
    themeColor: [
        { media: "(prefers-color-scheme: light)", color: "white" },
        { media: "(prefers-color-scheme: dark)", color: "black" },
    ],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
    return (
        <html suppressHydrationWarning lang="en">
            <head />
            <body className={clsx("min-h-screen bg-background font-sans antialiased", fontSans.variable)}>
                <Providers themeProps={{ attribute: "class", defaultTheme: "light" }}>
                    <div className="relative flex flex-col h-screen">
                        <div className="fixed top-0 left-0 right-0 z-50 bg-background">
                            <Navbar />
                        </div>
                        <main className="container pt-16 mx-auto max-w-7xl">{children}</main>
                        {/* <footer className="w-full flex items-center justify-center pb-2">
                            <Link
                                isExternal
                                className="flex items-center gap-1 text-current"
                                href="https://heroui.com?utm_source=next-app-template"
                                title="heroui.com homepage"
                            >
                                <span className="text-default-600">Powered by</span>
                                <p className="text-danger">HeroUI</p>
                            </Link>
                        </footer> */}
                    </div>
                </Providers>
            </body>
        </html>
    );
}
