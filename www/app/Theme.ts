/// <reference path="./References.d.ts"/>
import * as SuperAgent from 'superagent';
import * as Alert from './Alert';
import * as Csrf from './Csrf';
import * as MiscUtils from './utils/MiscUtils';
import * as EditorThemes from './EditorThemes';
import * as Monaco from "monaco-editor"

export interface Callback {
	(): void;
}

let callbacks: Set<Callback> = new Set<Callback>();
export let theme = 'dark';
export let themeVer = 5;
let editorThemeName = '';
export const monospaceSize = "12px"
export const monospaceFont = "Consolas, Menlo, 'Roboto Mono', 'DejaVu Sans Mono'"
export const monospaceWeight = "500"

export function save(): Promise<void> {
	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/theme')
			.send({
				theme: theme + `-${themeVer}`,
				editor_theme: editorThemeName,
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
					Alert.errorRes(res, 'Failed to save theme');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function themeVer3(): void {
  const blueprintTheme3 = document.getElementById(
    "blueprint3-theme") as HTMLLinkElement
  const blueprintTheme5 = document.getElementById(
    "blueprint5-theme") as HTMLLinkElement
  blueprintTheme3.disabled = false;
  blueprintTheme5.disabled = true;
	if (theme === "dark") {
		document.body.className = 'bp3-theme bp5-dark';
		document.documentElement.className = 'dark3-scroll';
	} else {
		document.body.className = 'bp3-theme';
		document.documentElement.className = '';
	}
  themeVer = 3;
}

export function themeVer5(): void {
  const blueprintTheme3 = document.getElementById(
    "blueprint3-theme") as HTMLLinkElement
  const blueprintTheme5 = document.getElementById(
    "blueprint5-theme") as HTMLLinkElement
  blueprintTheme3.disabled = true;
  blueprintTheme5.disabled = false;
	if (theme === "dark") {
		document.body.className = 'bp5-dark';
		document.documentElement.className = 'dark5-scroll';
	} else {
		document.body.className = '';
		document.documentElement.className = '';
	}
  themeVer = 5;
}

export function light(): void {
	theme = 'light';
	if (themeVer === 3) {
		document.body.className = 'bp3-theme';
		document.documentElement.className = '';
	} else {
		document.body.className = '';
		document.documentElement.className = '';
	}
	callbacks.forEach((callback: Callback): void => {
		callback();
	});
}

export function dark(): void {
	theme = 'dark';
	if (themeVer === 3) {
		document.body.className = 'bp3-theme bp5-dark';
		document.documentElement.className = 'dark3-scroll';
	} else {
		document.body.className = 'bp5-dark';
		document.documentElement.className = 'dark5-scroll';
	}
	callbacks.forEach((callback: Callback): void => {
		callback();
	});
}

export function toggle(): void {
  if (theme === "dark" && themeVer === 3) {
		light();
  } else if (theme === "light" && themeVer === 3) {
		dark();
    themeVer5();
  } else if (theme === "dark" && themeVer === 5) {
		light();
  } else if (theme === "light" && themeVer === 5) {
		dark();
    themeVer3();
  }
}

export function getEditorTheme(): string {
  if (!editorThemeName) {
    if (theme === "light") {
      return "github-light";
    } else {
      return "github-dark";
    }
  }
  return editorThemeName
}

export function setEditorTheme(name: string) {
	editorThemeName = name
	callbacks.forEach((callback: Callback): void => {
		callback();
	});
}

export function addChangeListener(callback: Callback): void {
	callbacks.add(callback);
}

export function removeChangeListener(callback: () => void): void {
	callbacks.delete(callback);
}

export let editorThemeNames: Record<string, string> = {}

for (let themeName in EditorThemes.editorThemes) {
	let editorTheme = EditorThemes.editorThemes[themeName]
	Monaco.editor.defineTheme(themeName, editorTheme)

	let formattedThemeName = MiscUtils.titleCase(
	themeName.replaceAll("-", " "))
	editorThemeNames[themeName] = formattedThemeName
}
