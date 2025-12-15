/// <reference path="../References.d.ts"/>
export const SYNC = 'storage.sync';
export const TRAVERSE = 'storage.traverse';
export const FILTER = 'storage.filter';
export const CHANGE = 'storage.change';

export interface Storage {
	id?: string;
	name?: string;
	comment?: string;
	type?: string;
	endpoint?: string;
	bucket?: string;
	access_key?: string;
	secret_key?: string;
	insecure?: boolean;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
}

export type Storages = Storage[];

export type StorageRo = Readonly<Storage>;
export type StoragesRo = ReadonlyArray<StorageRo>;

export interface StorageDispatch {
	type: string;
	data?: {
		id?: string;
		storage?: Storage;
		storages?: Storages;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
