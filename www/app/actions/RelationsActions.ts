/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import * as RelationTypes from '../types/RelationTypes';
import OrganizationsStore from '../stores/OrganizationsStore';

export function load(kind: string,
	id: string): Promise<RelationTypes.Relation> {

	return new Promise<RelationTypes.Relation>((resolve, reject): void => {
		SuperAgent
			.get("/relations/" + kind + "/" + id)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve(null);
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load resource overview');
					reject(err);
					return;
				}

				resolve(res.body as RelationTypes.Relation);
			});
	});
}
