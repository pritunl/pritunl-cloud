/// <reference path="../References.d.ts"/>
export const SYNC = 'block.sync';
export const TRAVERSE = 'block.traverse';
export const FILTER = 'block.filter';
export const CHANGE = 'block.change';

export interface Block {
	id?: string;
	name?: string;
	comment?: string;
	type?: string;
	subnets?: string[];
	subnets6?: string[];
	excludes?: string[];
	netmask?: string;
	gateway?: string;
	gateway6?: string;
}

export type Blocks = Block[];

export type BlockRo = Readonly<Block>;
export type BlocksRo = ReadonlyArray<BlockRo>;

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
}

export interface BlockDispatch {
	type: string;
	data?: {
		id?: string;
		block?: Block;
		blocks?: Blocks;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
