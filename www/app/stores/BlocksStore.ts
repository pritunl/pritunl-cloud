/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as BlockTypes from '../types/BlockTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class BlocksStore extends EventEmitter {
	_blocks: BlockTypes.BlocksRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._blocks = Object.freeze([]);
		this._map = {};
		this.emitChange();
	}

	get blocks(): BlockTypes.BlocksRo {
		return this._blocks;
	}

	get blocksM(): BlockTypes.Blocks {
		let blocks: BlockTypes.Blocks = [];
		this._blocks.forEach((
				block: BlockTypes.BlockRo): void => {
			blocks.push({
				...block,
			});
		});
		return blocks;
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

	_sync(blocks: BlockTypes.Block[]): void {
		this._map = {};
		for (let i = 0; i < blocks.length; i++) {
			blocks[i] = Object.freeze(blocks[i]);
			this._map[blocks[i].id] = i;
		}

		this._blocks = Object.freeze(blocks);
		this.emitChange();
	}

	_callback(action: BlockTypes.BlockDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case BlockTypes.SYNC:
				this._sync(action.data.blocks);
				break;
		}
	}
}

export default new BlocksStore();
