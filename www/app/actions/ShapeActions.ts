/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as ShapeTypes from '../types/ShapeTypes';
import ShapesStore from '../stores/ShapesStore';
import * as MiscUtils from '../utils/MiscUtils';

let syncId: string;

export function sync(noLoading?: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/shape')
			.query({
				...ShapesStore.filter,
				page: ShapesStore.page,
				page_count: ShapesStore.pageCount,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				if (loader) {
					loader.done();
				}

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (curSyncId !== syncId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load shapes');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: ShapeTypes.SYNC,
					data: {
						shapes: res.body.shapes,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: ShapeTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: ShapeTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: ShapeTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(shape: ShapeTypes.Shape): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/shape/' + shape.id)
			.send(shape)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to save shape');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(shape: ShapeTypes.Shape): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/shape')
			.send(shape)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to create shape');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(shapeId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/shape/' + shapeId)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete shape');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(shapeIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/shape')
			.send(shapeIds)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete shapes');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: ShapeTypes.ShapeDispatch) => {
	switch (action.type) {
		case ShapeTypes.CHANGE:
			sync();
			break;
	}
});
