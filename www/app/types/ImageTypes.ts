/// <reference path="../References.d.ts"/>
export const SYNC = 'image.sync';
export const SYNC_DATACENTER = 'image.sync_datacenter';
export const TRAVERSE = 'image.traverse';
export const FILTER = 'image.filter';
export const CHANGE = 'image.change';

export interface Image {
	id: string;
	name?: string;
	organization?: string;
	storage?: string;
	key?: string;
	type?: string;
	etag?: string;
	last_modified?: string;
}

export interface Filter {
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
