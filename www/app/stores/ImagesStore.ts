/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as ImageTypes from '../types/ImageTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class ImagesStore extends EventEmitter {
	_images: ImageTypes.ImagesRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: ImageTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._images = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get images(): ImageTypes.ImagesRo {
		return this._images;
	}

	get imagesM(): ImageTypes.Images {
		let images: ImageTypes.Images = [];
		this._images.forEach((image: ImageTypes.ImageRo): void => {
			images.push({
				...image,
			});
		});
		return images;
	}

	get page(): number {
		return this._page || 0;
	}

	get pageCount(): number {
		return this._pageCount || 20;
	}

	get pages(): number {
		return Math.ceil(this.count / this.pageCount);
	}

	get filter(): ImageTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	image(id: string): ImageTypes.ImageRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._images[i];
	}

	emitChange(): void {
		this.emitDefer(GlobalTypes.CHANGE);
	}

	addChangeListener(callback: () => void): void {
		this.on(GlobalTypes.CHANGE, callback);
	}

	removeChangeListener(callback: () => void): void {
		this.removeListener(GlobalTypes.CHANGE, callback);
	}

	_traverse(page: number): void {
		this._page = Math.min(this.pages, page);
	}

	_filterCallback(filter: ImageTypes.Filter): void {
		if ((this._filter !== null && filter === null) ||
			(!Object.keys(this._filter || {}).length && filter !== null) || (
				filter && this._filter && (
					filter.name !== this._filter.name
				))) {
			this._traverse(0);
		}
		this._filter = filter;
		this.emitChange();
	}

	_sync(images: ImageTypes.Image[], count: number): void {
		this._map = {};
		for (let i = 0; i < images.length; i++) {
			images[i] = Object.freeze(images[i]);
			this._map[images[i].id] = i;
		}

		this._count = count;
		this._images = Object.freeze(images);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: ImageTypes.ImageDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case ImageTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case ImageTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case ImageTypes.SYNC:
				this._sync(action.data.images, action.data.count);
				break;
		}
	}
}

export default new ImagesStore();
