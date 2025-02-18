/* eslint-disable no-console */
export async function fetchCourses(page: number = 1, pageSize: number = 50, query: string = "*") {
    try {
        const response = await fetch(
            // `https://badger-class-tracker-backend.vercel.app/api/courses?page=${page}&pageSize=${pageSize}&query=${query}`,
            `http://localhost:8000/api/courses?page=${page}&pageSize=${pageSize}&query=${query}`,
            {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                },
            }
        );

        if (!response.ok) {
            console.error(`❌ API request failed with status: ${response.status}`);

            return { hits: [], found: 0 };
        }

        const data = await response.json();

        if (!data.hits) {
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
        console.error("❌ Error fetching courses:", error);

        return { hits: [], found: 0 };
    }
}
