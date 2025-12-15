/// <reference path="../References.d.ts"/>
export const SYNC = 'policy.sync';
export const TRAVERSE = 'policy.traverse';
export const FILTER = 'policy.filter';
export const CHANGE = 'policy.change';

export interface Rule {
	type?: string;
	disable?: boolean;
	values?: string[];
}

export interface Policy {
	id?: string;
	name?: string;
	comment?: string;
	disabled?: boolean;
	roles?: string[];
	rules?: {[key: string]: Rule};
	admin_secondary?: string;
	user_secondary?: string;
	admin_device_secondary?: boolean;
	user_device_secondary?: boolean;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
}

export type Policies = Policy[];

export type PolicyRo = Readonly<Policy>;
export type PoliciesRo = ReadonlyArray<PolicyRo>;

export interface PolicyDispatch {
	type: string;
	data?: {
		id?: string;
		policy?: Policy;
		policies?: Policies;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
