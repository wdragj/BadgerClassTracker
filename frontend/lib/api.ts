export async function fetchCourses(page: number = 1, pageSize: number = 50, query: string = "*") {
    try {
        const response = await fetch(
            `${process.env.NEXT_PUBLIC_BACKEND_URL}/api/courses?page=${page}&pageSize=${pageSize}&query=${encodeURIComponent(query)}`,
            {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                },
            }
        );

        if (!response.ok) {
            // eslint-disable-next-line no-console
            console.error(`❌ API request failed with status: ${response.status}`);

            return { hits: [], found: 0 };
        }

        const data = await response.json();

        if (!data.hits) {
            // eslint-disable-next-line no-console
            console.warn("⚠️ No courses found.");

            return { hits: [], found: 0 };
        }

        return {
            hits: data.hits.map((course: any) => ({
                id: course.courseId,
                name: course.courseDesignation,
                fullname: course.fullCourseDesignation,
                title: course.title,
                subject: course.subject.shortDescription,
                subjectCode: course.subject.subjectCode,
                termCode: course.subject.termCode,
                credits: course.creditRange,
                description: course.description,
                enrollmentPrerequisites: course.enrollmentPrerequisites || "None",
                typicallyOffered: course.typicallyOffered || "N/A",
                repeatable: course.repeatable === "Y" ? "Yes" : "No",
            })),
            found: data.found,
        };
    } catch (error) {
        // eslint-disable-next-line no-console
        console.error("❌ Error fetching courses:", error);

        return { hits: [], found: 0 };
    }
}
