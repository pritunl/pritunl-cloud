/// <reference path="../References.d.ts"/>

export interface Kind {
	name: string
	label: string
	title: string
}

export interface Resource {
	name: string
	info: ResourceInfo[]
}

export interface ResourceInfo {
	label: string
	value: string
}

export interface Dispatch {
	type: string
}
