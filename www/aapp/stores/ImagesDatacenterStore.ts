/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as ImageTypes from '../types/ImageTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class ImagesDatacenterStore extends EventEmitter {
	_images: ImageTypes.ImagesRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	get images(): ImageTypes.ImagesRo {
		return this._images;
	}

	get imagesM(): ImageTypes.Images {
		let images: ImageTypes.Images = [];
		this._images.forEach((
				image: ImageTypes.ImageRo): void => {
			images.push({
				...image,
			});
		});
		return images;
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

	_sync(images: ImageTypes.Image[]): void {
		this._map = {};
		for (let i = 0; i < images.length; i++) {
			images[i] = Object.freeze(images[i]);
			this._map[images[i].id] = i;
		}

		this._images = Object.freeze(images);
		this.emitChange();
	}

	_callback(action: ImageTypes.ImageDispatch): void {
		switch (action.type) {
			case ImageTypes.SYNC_DATACENTER:
				this._sync(action.data.images);
				break;
		}
	}
}

export default new ImagesDatacenterStore();
