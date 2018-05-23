/// <reference path="./References.d.ts"/>
import * as SuperAgent from 'superagent';
import * as Alert from './Alert';
import * as Csrf from './Csrf';

export let oracle = false;

export function save(): Promise<void> {
	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/license')
			.send({
				oracle: oracle,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to save license state');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function setOracle(state: boolean): void {
	oracle = state;
}
