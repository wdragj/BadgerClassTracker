"use client";

import type { Course } from "@/app/page";

import React from "react";
import { Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, Button } from "@heroui/react";

interface UnsubscribeModalProps {
    isOpen: boolean;
    course: Course;
    onClose: () => void;
    onUnsubscribe: (course: Course) => void;
}

export default function UnsubscribeModal({ isOpen, course, onClose, onUnsubscribe }: UnsubscribeModalProps) {
    return (
        <Modal isOpen={isOpen} placement="center" size="xs" onClose={onClose}>
            <ModalContent>
                {(close) => (
                    <>
                        <ModalHeader>Unsubscribe</ModalHeader>
                        <ModalBody>
                            <p>
                                Are you sure you want to unsubscribe from <strong>{course.name}</strong>?
                            </p>
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
                                    onUnsubscribe(course);
                                    close();
                                }}
                            >
                                Unsubscribe
                            </Button>
                        </ModalFooter>
                    </>
                )}
            </ModalContent>
        </Modal>
    );
}
