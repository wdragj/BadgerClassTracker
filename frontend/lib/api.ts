/* eslint-disable no-console */
export async function fetchCourses(query: string = "*") {
    try {
        const response = await fetch(`https://badger-class-tracker-backend.vercel.app/api/courses`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
        });

        if (!response.ok) {
            console.error(`❌ API request failed with status: ${response.status}`);

            return [];
        }

        const data = await response.json();

        if (!data.hits) {
            console.warn("⚠️ No courses found.");

            return [];
        }

        return data.hits.map((course: any) => ({
            id: course.courseId,
            name: course.courseDesignation, // e.g., "COMP SCI 403"
            fullname: course.fullCourseDesignation, // e.g., "COMPUTER SCIENCES 403"
            title: course.title, // e.g., "INTRODUCTION TO COMPUTER SCIENCE II"
            subject: course.subject.shortDescription, // e.g., "COMP SCI"
            termCode: course.subject.termCode, // e.g., "1254"
            credits: course.creditRange, // e.g., "1"
            description: course.description, // Course description
            enrollmentPrerequisites: course.enrollmentPrerequisites || "None",
            typicallyOffered: course.typicallyOffered || "N/A",
            repeatable: course.repeatable === "Y" ? "Yes" : "No",
        }));
    } catch (error) {
        console.error("❌ Error fetching courses:", error);

        return [];
    }
}
