/// <reference path="../References.d.ts"/>
export const SYNC = 'storage.sync';
export const CHANGE = 'storage.change';

export interface Storage {
	id: string;
	name?: string;
	type?: string;
	endpoint?: string;
	bucket?: string;
	access_key?: string;
	secret_key?: string;
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
	};
}
