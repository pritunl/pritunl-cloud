/// <reference path="../References.d.ts"/>
export const SYNC = 'image.sync';
export const SYNC_DATACENTER = 'image.sync_datacenter';
export const TRAVERSE = 'image.traverse';
export const FILTER = 'image.filter';
export const CHANGE = 'image.change';

export interface Image {
	id?: string;
	disk_id?: string;
	name?: string;
	release?: string;
	build?: string;
	comment?: string;
	organization?: string;
	storage?: string;
	signed?: boolean;
	key?: string;
	type?: string;
	firmware?: string;
	etag?: string;
	last_modified?: string;
	storage_class?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	type?: string;
	organization?: string;
}

export type Images = Image[];

export type ImageRo = Readonly<Image>;
export type ImagesRo = ReadonlyArray<ImageRo>;

export interface ImageDispatch {
	type: string;
	data?: {
		id?: string;
		image?: Image;
		images?: Images;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
