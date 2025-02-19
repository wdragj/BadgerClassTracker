"use client";

import React from "react";
import { Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, Button } from "@heroui/react";
import type { Course } from "@/app/page";

interface SubscribeModalProps {
    isOpen: boolean;
    course: Course;
    onClose: () => void;
    onSubscribe: (course: Course) => void;
}

export default function SubscribeModal({ isOpen, course, onClose, onSubscribe }: SubscribeModalProps) {
    return (
        <Modal isOpen={isOpen} size="xs" onClose={onClose} placement="center">
            <ModalContent>
                {(close) => (
                    <>
                        <ModalHeader>Subscribe</ModalHeader>
                        <ModalBody>
                            <p>
                                Are you sure you want to subscribe to <strong>{course.name}</strong>?
                            </p>
                            {/* <p>Title: {course.title}</p>
                            <p>Credits: {course.credits}</p> */}
                        </ModalBody>
                        <ModalFooter>
                            <Button variant="flat" color="danger" onPress={close}>
                                Cancel
                            </Button>
                            <Button
                                fullWidth
                                variant="flat"
                                color="primary"
                                onPress={() => {
                                    onSubscribe(course);
                                    close();
                                }}
                            >
                                Subscribe
                            </Button>
                        </ModalFooter>
                    </>
                )}
            </ModalContent>
        </Modal>
    );
}
