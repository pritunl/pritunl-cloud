/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as BlockTypes from '../types/BlockTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class BlocksStore extends EventEmitter {
	_blocks: BlockTypes.BlocksRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: BlockTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._blocks = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get blocks(): BlockTypes.BlocksRo {
		return this._blocks;
	}

	get blocksM(): BlockTypes.Blocks {
		let blocks: BlockTypes.Blocks = [];
		this._blocks.forEach((block: BlockTypes.BlockRo): void => {
			blocks.push({
				...block,
			});
		});
		return blocks;
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

	get filter(): BlockTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	block(id: string): BlockTypes.BlockRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._blocks[i];
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

	_filterCallback(filter: BlockTypes.Filter): void {
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

	_sync(blocks: BlockTypes.Block[], count: number): void {
		this._map = {};
		for (let i = 0; i < blocks.length; i++) {
			blocks[i] = Object.freeze(blocks[i]);
			this._map[blocks[i].id] = i;
		}

		this._count = count;
		this._blocks = Object.freeze(blocks);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: BlockTypes.BlockDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case BlockTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case BlockTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case BlockTypes.SYNC:
				this._sync(action.data.blocks, action.data.count);
				break;
		}
	}
}

export default new BlocksStore();
