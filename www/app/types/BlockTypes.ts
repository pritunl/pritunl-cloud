/// <reference path="../References.d.ts"/>
export const SYNC = 'block.sync';
export const CHANGE = 'block.change';

export interface Block {
	id: string;
	name?: string;
	addresses?: string[];
	excludes?: string[];
	netmask?: string;
	gateway?: string;
}

export type Blocks = Block[];

export type BlockRo = Readonly<Block>;
export type BlocksRo = ReadonlyArray<BlockRo>;

export interface BlockDispatch {
	type: string;
	data?: {
		id?: string;
		block?: Block;
		blocks?: Blocks;
	};
}
