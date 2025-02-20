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

            return { term: null, hits: [], found: 0 };
        }

        const data = await response.json();

        // Ensure that the courses property exists and has hits
        if (!data.courses || !data.courses.hits) {
            // eslint-disable-next-line no-console
            console.warn("⚠️ No courses found.");

            return { term: data.term || null, hits: [], found: 0 };
        }

        return {
            term: data.term,
            hits: data.courses.hits.map((course: any) => ({
                id: course.courseId,
                name: course.courseDesignation,
                fullname: course.fullCourseDesignation,
                title: course.title,
                subject: course.subject?.longDescription,
                subjectCode: course.subject?.subjectCode,
                termCode: course.subject?.termCode,
                credits: course.creditRange,
                description: course.description,
                enrollmentPrerequisites: course.enrollmentPrerequisites || "None",
                typicallyOffered: course.typicallyOffered || "N/A",
                repeatable: course.repeatable === "Y" ? "Yes" : "No",
            })),
            found: data.courses.found,
        };
    } catch (error) {
        // eslint-disable-next-line no-console
        console.error("❌ Error fetching courses:", error);

        return { term: null, hits: [], found: 0 };
    }
}
