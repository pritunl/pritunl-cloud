/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PodTypes from '../types/PodTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import PodsStore from '../stores/PodsStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as PodActions from '../actions/PodActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import Pod from './Pod';
import PodDetailed from './PodDetailed';
import PodNew from './PodNew';
import PodsFilter from './PodsFilter';
import PodsPage from './PodsPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import CompletionStore from "../stores/CompletionStore";
import CompletionCache from "../completion/Cache"
import * as CompletionTypes from "../types/CompletionTypes";
import * as CompletionActions from "../actions/CompletionActions";

interface Selected {
	[key: string]: boolean;
}

interface State {
	podId: string;
	pods: PodTypes.PodsRo;
	filter: PodTypes.Filter;
	organizations: OrganizationTypes.OrganizationsRo;
	completion: CompletionTypes.Completion;
	selected: Selected;
	newOpened: boolean;
	lastSelected: string;
	disabled: boolean;
	sidebar: boolean;
	settings: boolean;
	mode: string;
}

const css = {
	layout: {
		height: 'calc(100dvh - 230px)',
		maxHeight: 'calc(100dvh - 230px)',
	} as React.CSSProperties,
	items: {
		width: '200px',
		borderSpacing: '0 5px',
		padding: '0 5px'
	} as React.CSSProperties,
	itemsBox: {
		width: '100%',
		overflowY: 'auto',
	} as React.CSSProperties,
	pods: {
		flexGrow: 1,
		minWidth: 0,
		maxWidth: '100%',
	} as React.CSSProperties,
	placeholder: {
		opacity: 0,
		width: '100%',
	} as React.CSSProperties,
	header: {
		marginTop: '-19px',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
	headingBeta: {
		position: "relative",
		top: "-4px",
		left: "6px",
	} as React.CSSProperties,
	button: {
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
	nonState: {
		justifyContent: 'center',
		height: '100%',
	} as React.CSSProperties,
};

export default class Pods extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			podId: null,
			pods: PodsStore.pods,
			filter: PodsStore.filter,
			organizations: OrganizationsStore.organizations,
			completion: CompletionStore.completion,
			selected: {},
			newOpened: false,
			lastSelected: null,
			disabled: false,
			sidebar: true,
			settings: false,
			mode: "view",
		};
		this.handleKeyDown = this.handleKeyDown.bind(this);
	}

	get selected(): boolean {
		return !!Object.keys(this.state.selected).length;
	}

	componentDidMount(): void {
		PodsStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		PodActions.sync();
		OrganizationActions.sync();
		CompletionActions.sync();
		document.addEventListener('keydown', this.handleKeyDown);
	}

	componentWillUnmount(): void {
		PodsStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
		document.removeEventListener('keydown', this.handleKeyDown);
	}

	handleKeyDown(event: KeyboardEvent) {
		if (event.ctrlKey && event.key === "e") {
			this.setState({
				...this.state,
				sidebar: !this.state.sidebar,
			})
			event.preventDefault();
		}
	}

	onChange = (): void => {
		let pods = PodsStore.pods;
		let selected: Selected = {};
		let curSelected = this.state.selected;

		pods.forEach((pod: PodTypes.Pod): void => {
			if (curSelected[pod.id]) {
				selected[pod.id] = true;
			}
		});

		CompletionCache.update(CompletionStore.completion)

		this.setState({
			...this.state,
			pods: PodsStore.pods,
			filter: PodsStore.filter,
			organizations: OrganizationsStore.organizations,
			selected: selected,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.removeMulti(
				Object.keys(this.state.selected)).then((): void => {
			this.setState({
				...this.state,
				selected: {},
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let activePod: PodTypes.PodRo;
		let podsDom: JSX.Element[] = [];

		this.state.pods.forEach((
				pod: PodTypes.PodRo): void => {
			if (pod.id === this.state.podId) {
				activePod = pod
			}

			podsDom.push(<Pod
				key={pod.id}
				pod={pod}
				organizations={this.state.organizations}
				selected={!!this.state.selected[pod.id]}
				open={activePod?.id === pod.id}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let pods = this.state.pods;
						let start: number;
						let end: number;

						for (let i = 0; i < pods.length; i++) {
							let usr = pods[i];

							if (usr.id === pod.id) {
								start = i;
							} else if (usr.id === this.state.lastSelected) {
								end = i;
							}
						}

						if (start !== undefined && end !== undefined) {
							if (start > end) {
								end = [start, start = end][0];
							}

							for (let i = start; i <= end; i++) {
								selected[pods[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: pod.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[pod.id]) {
						delete selected[pod.id];
					} else {
						selected[pod.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: pod.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let newMode = this.state.mode
					if (this.state.mode === "edit") {
						newMode = "view"
					}
					this.setState({
						...this.state,
						podId: pod.id,
						newOpened: false,
						mode: newMode,
					});
				}}
			/>);
		});

		let podElem: JSX.Element;
		if (this.state.newOpened) {
			podElem = <PodNew
				organizations={this.state.organizations}
				onClose={(): void => {
					this.setState({
						...this.state,
						newOpened: false,
					});
				}}
				sidebar={this.state.sidebar}
				toggleSidebar={() => {
					this.setState({
						...this.state,
						sidebar: !this.state.sidebar,
					})
				}}
			/>;
		} else if (!podsDom.length) {
			podElem = <div className="layout vertical" style={css.nonState}>
				<NonState
					hidden={false}
					iconClass="bp5-icon-cube"
					title="No pods"
					description="Add a new pod to get started."
				/>
			</div>
		} else if (activePod) {
			podElem = <PodDetailed
				key={activePod?.id}
				organizations={this.state.organizations}
				pod={activePod}
				mode={this.state.mode}
				onMode={(mode: string) => {
					this.setState({
						...this.state,
						mode: mode,
					})
				}}
				settings={this.state.settings}
				toggleSettings={() => {
					this.setState({
						...this.state,
						settings: !this.state.settings,
					})
				}}
				sidebar={this.state.sidebar}
				toggleSidebar={() => {
					this.setState({
						...this.state,
						sidebar: !this.state.sidebar,
					})
				}}
			/>
		}

		let filterClass = 'bp5-button bp5-intent-primary bp5-icon-filter ';
		if (this.state.filter) {
			filterClass += 'bp5-active';
		}

		let selectedNames: string[] = [];
		for (let instId of Object.keys(this.state.selected)) {
			let inst = PodsStore.pod(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		return <Page full={true}>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Pods
						<span
							style={css.headingBeta}
							className="bp5-tag bp5-intent-primary"
						>Beta</span>
					</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className="bp5-button bp5-icon-help"
							style={css.button}
							type="button"
							onClick={(): void => {
								console.log("TODO")
							}}
						>
							Help
						</button>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									PodActions.filter({});
								} else {
									PodActions.filter(null);
								}
							}}
						>
							Filters
						</button>
						<ConfirmButton
							label="Delete Selected"
							className="bp5-intent-danger bp5-icon-delete"
							progressClassName="bp5-intent-danger"
							safe={true}
							style={css.button}
							confirmMsg="Permanently delete the selected pods"
							confirmInput={true}
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="bp5-button bp5-intent-success bp5-icon-add"
							style={css.button}
							disabled={this.state.disabled || this.state.newOpened}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									newOpened: true,
								});
							}}
						>New</button>
					</div>
				</div>
			</PageHeader>
			<PodsFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					PodActions.filter(filter);
				}}
				organizations={this.state.organizations}
			/>
			<div className="layout horizontal" style={css.layout}>
				<div
					className="bp5-menu"
					style={css.items}
					hidden={!this.state.sidebar}
				>
					{podsDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={1} style={css.placeholder}/>
					</tr>
				</div>
				<div style={css.pods}>
					{podElem}
				</div>
			</div>
			<PodsPage
				onPage={(): void => {
					this.setState({
						...this.state,
						selected: {},
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}
