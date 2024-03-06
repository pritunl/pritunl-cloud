/// <reference path="../References.d.ts"/>
export const SYNC = 'shape.sync';
export const TRAVERSE = 'shape.traverse';
export const FILTER = 'shape.filter';
export const CHANGE = 'shape.change';

export interface Shape {
	id?: string;
	name?: string;
	comment?: string;
	type?: string;
	delete_protection?: boolean;
	zone?: string;
	roles?: string[];
	flexible?: boolean;
	disk_type?: string;
	disk_pool?: string;
	memory?: number;
	processors?: number;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
	role?: string;
}

export type Shapes = Shape[];

export type ShapeRo = Readonly<Shape>;
export type ShapesRo = ReadonlyArray<ShapeRo>;

export interface ShapeDispatch {
	type: string;
	data?: {
		id?: string;
		shape?: Shape;
		shapes?: Shapes;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
