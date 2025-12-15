/// <reference path="../References.d.ts"/>
export const SYNC = 'plan.sync';
export const SYNC_NAME = 'plan.sync_name';
export const TRAVERSE = 'plan.traverse';
export const FILTER = 'plan.filter';
export const CHANGE = 'plan.change';

export interface Plan {
	id?: string;
	name?: string;
	comment?: string;
	organization?: string;
	type?: string;
	statements?: Statement[];
}

export interface Statement {
	id?: string;
	statement?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	organization?: string;
}

export type Plans = Plan[];

export type PlanRo = Readonly<Plan>;
export type PlansRo = ReadonlyArray<PlanRo>;

export interface PlanDispatch {
	type: string;
	data?: {
		id?: string;
		plan?: Plan;
		plans?: Plans;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
