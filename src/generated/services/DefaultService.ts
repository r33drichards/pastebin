/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CompletionResponse } from '../models/CompletionResponse';
import type { Diff } from '../models/Diff';
import type { Paste } from '../models/Paste';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * Create a new paste
     * @param formData
     * @returns void
     * @throws ApiError
     */
    public static createPaste(
        formData: {
            /**
             * The text content of the paste
             */
            text: string;
            /**
             * The programming language for syntax highlighting
             */
            lang: string;
        },
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/paste',
            formData: formData,
            mediaType: 'application/x-www-form-urlencoded',
            errors: {
                302: `Paste created successfully`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get a paste by ID
     * @param id The paste ID
     * @returns Paste Paste retrieved successfully
     * @throws ApiError
     */
    public static getPaste(
        id: string,
    ): CancelablePromise<Paste> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/paste',
            query: {
                'id': id,
            },
            errors: {
                404: `Paste not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Create a new diff
     * @param formData
     * @returns void
     * @throws ApiError
     */
    public static createDiff(
        formData: {
            /**
             * The original text
             */
            original: string;
            /**
             * The modified text
             */
            modified: string;
        },
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/diff',
            formData: formData,
            mediaType: 'application/x-www-form-urlencoded',
            errors: {
                302: `Diff created successfully`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get a diff by ID
     * @param id The diff ID
     * @returns Diff Diff retrieved successfully
     * @throws ApiError
     */
    public static getDiff(
        id: string,
    ): CancelablePromise<Diff> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/diff',
            query: {
                'id': id,
            },
            errors: {
                404: `Diff not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get code completion suggestions
     * @param formData
     * @returns CompletionResponse Completions retrieved successfully
     * @throws ApiError
     */
    public static getCompletion(
        formData: {
            /**
             * The text to get completions for
             */
            text: string;
        },
    ): CancelablePromise<CompletionResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/complete',
            formData: formData,
            mediaType: 'application/x-www-form-urlencoded',
            errors: {
                500: `Internal server error`,
            },
        });
    }
    /**
     * Health check
     * @returns string Service is healthy
     * @throws ApiError
     */
    public static healthCheck(): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/health',
        });
    }
}
