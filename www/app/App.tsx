/// <reference path="References.d.ts"/>
import * as Monaco from "monaco-editor";
import * as MonacoEditor from "@monaco-editor/react";

MonacoEditor.loader.config({
	monaco: Monaco
})

import * as React from 'react';
import * as ReactDOM from 'react-dom';
import * as Blueprint from '@blueprintjs/core';
import Main from './components/Main';
import * as Alert from './Alert';
import * as Event from './Event';
import * as Csrf from './Csrf';
import * as MiscUtils from './utils/MiscUtils';
import * as CompletionActions from './actions/CompletionActions';

import hljs from 'highlight.js/lib/core';
import plaintext from 'highlight.js/lib/languages/plaintext';
import bash from 'highlight.js/lib/languages/bash';
import python from 'highlight.js/lib/languages/python';
import yaml from 'highlight.js/lib/languages/yaml';

Csrf.load().then((): void => {
	Blueprint.FocusStyleManager.onlyShowFocusOnTabs();
	Alert.init();
	Event.init();

	new MiscUtils.SyncInterval(
		async () => {
			let lastSync = CompletionActions.lastSync()
			if (lastSync && (Date.now() - lastSync) > 5000) {
				CompletionActions.sync();
			}
		},
		1000,
	)

	hljs.registerLanguage('plaintext', plaintext)
	hljs.registerLanguage('shell', bash)
	hljs.registerLanguage('python', python)
	hljs.registerLanguage('yaml', yaml)

	ReactDOM.render(
		<Blueprint.OverlaysProvider>
			<div><Main/></div>
		</Blueprint.OverlaysProvider>,
		document.getElementById('app'),
	);
});
