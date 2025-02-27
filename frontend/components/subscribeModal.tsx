"use client";

import type { Course } from "@/app/page";

import React from "react";
import { Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, Button } from "@heroui/react";

interface SubscribeModalProps {
    isOpen: boolean;
    course: Course;
    onClose: () => void;
    onSubscribe: (course: Course) => void;
}

export default function SubscribeModal({ isOpen, course, onClose, onSubscribe }: SubscribeModalProps) {
    return (
        <Modal isOpen={isOpen} placement="center" size="xs" onClose={onClose}>
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
                            <Button color="danger" variant="flat" onPress={close}>
                                Cancel
                            </Button>
                            <Button
                                fullWidth
                                color="primary"
                                variant="flat"
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
