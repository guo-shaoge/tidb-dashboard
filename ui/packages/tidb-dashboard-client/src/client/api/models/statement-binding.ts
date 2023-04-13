/* tslint:disable */
/* eslint-disable */
/**
 * Dashboard API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 1.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */



/**
 * 
 * @export
 * @interface StatementBinding
 */
export interface StatementBinding {
    /**
     * 
     * @type {string}
     * @memberof StatementBinding
     */
    'plan_digest'?: string;
    /**
     * 
     * @type {string}
     * @memberof StatementBinding
     */
    'source'?: StatementBindingSourceEnum;
    /**
     * 
     * @type {string}
     * @memberof StatementBinding
     */
    'status'?: StatementBindingStatusEnum;
}

export const StatementBindingSourceEnum = {
    manual: 'manual',
    history: 'history',
    capture: 'capture',
    evolve: 'evolve'
} as const;

export type StatementBindingSourceEnum = typeof StatementBindingSourceEnum[keyof typeof StatementBindingSourceEnum];
export const StatementBindingStatusEnum = {
    enabled: 'enabled',
    using: 'using',
    disabled: 'disabled',
    deleted: 'deleted',
    invalid: 'invalid',
    rejected: 'rejected',
    pending_verify: 'pending verify'
} as const;

export type StatementBindingStatusEnum = typeof StatementBindingStatusEnum[keyof typeof StatementBindingStatusEnum];

