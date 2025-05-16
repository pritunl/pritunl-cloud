/// <reference path="./References.d.ts"/>
import * as SuperAgent from 'superagent';
import * as License from './License';
import * as Theme from './Theme';

export let token = '';

export function load(): Promise<void> {
	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/csrf')
			.set('Accept', 'application/json')
			.end((err: any, res: SuperAgent.Response): void => {
				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					reject(err);
					return;
				}

				token = res.body.token;

				License.setOracle(!!res.body.oracle_license);

				let theme = res.body.theme
				if (theme) {
					let themeParts = theme.split("-")
					if (themeParts[1] === "3") {
						Theme.themeVer3()
					} else {
						Theme.themeVer5()
					}

					if (themeParts[0] === "light") {
						Theme.light();
					} else {
						Theme.dark();
					}
				} else {
					Theme.dark();
				}

				if (res.body.editor_theme) {
					Theme.setEditorTheme(res.body.editor_theme);
				}

				resolve();
			});
	});
}
