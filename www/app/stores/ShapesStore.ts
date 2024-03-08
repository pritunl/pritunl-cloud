/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as ShapeTypes from '../types/ShapeTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class ShapesStore extends EventEmitter {
	_shapes: ShapeTypes.ShapesRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: ShapeTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._shapes = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get shapes(): ShapeTypes.ShapesRo {
		return this._shapes;
	}

	get shapesM(): ShapeTypes.Shapes {
		let shapes: ShapeTypes.Shapes = [];
		this._shapes.forEach((shape: ShapeTypes.ShapeRo): void => {
			shapes.push({
				...shape,
			});
		});
		return shapes;
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

	get filter(): ShapeTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	shape(id: string): ShapeTypes.ShapeRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._shapes[i];
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

	_filterCallback(filter: ShapeTypes.Filter): void {
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

	_sync(shapes: ShapeTypes.Shape[], count: number): void {
		this._map = {};
		for (let i = 0; i < shapes.length; i++) {
			shapes[i] = Object.freeze(shapes[i]);
			this._map[shapes[i].id] = i;
		}

		this._count = count;
		this._shapes = Object.freeze(shapes);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: ShapeTypes.ShapeDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case ShapeTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case ShapeTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case ShapeTypes.SYNC:
				this._sync(action.data.shapes, action.data.count);
				break;
		}
	}
}

export default new ShapesStore();
