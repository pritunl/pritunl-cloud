/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Blueprint from "@blueprintjs/core"
import * as Theme from '../Theme';
import * as PodTypes from "../types/PodTypes"
import * as PodActions from "../actions/PodActions"
import PageInput from './PageInput';
import PageSave from './PageSave';
import PageInputButton from './PageInputButton';
import Help from "./Help"

interface Props {
	disabled: boolean
	commitMap: Record<string, number>
	deployment: PodTypes.Deployment
	open: boolean
	onClose: () => void
}

interface State {
	disabled: boolean
	changed: boolean
	message: string
	deployment: PodTypes.Deployment
	addTag: string
}

const css = {
	container: {
		height: "900px",
		overflowY: "auto",
		marginBottom: "10px",
	} as React.CSSProperties,
	card: {
		padding: "5px 5px 3px 5px",
	} as React.CSSProperties,
	cardInactive: {
		padding: "5px 5px 3px 5px",
		opacity: 0.6,
	} as React.CSSProperties,
	check: {
		margin: "0 0 0 0",
	} as React.CSSProperties,
	checkBox: {
		display: "flex",
		paddingBottom: "2px",
	} as React.CSSProperties,
	info: {
		marginBottom: "0px",
	} as React.CSSProperties,
	settings: {
		padding: "0 7px",
		marginTop: "12px",
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	itemFirst: {
		flex: "1 1 auto",
		minWidth: "100px",
		margin: "0 5px 0 -4px",
	} as React.CSSProperties,
	item: {
		flex: "1 1 auto",
		minWidth: "30px",
		margin: " 0 5px",
	} as React.CSSProperties,
	itemLast: {
		flex: "0 1 auto",
		minWidth: "30px",
		margin: " 0 5px",
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
}

export default class PodDeploymentEdit extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
			disabled: false,
			changed: false,
			message: "",
			deployment: null,
			addTag: "",
		}
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.commitDeployment(this.state.deployment).then((): void => {
			this.setState({
				...this.state,
				message: 'Your changes have been saved',
				changed: false,
				disabled: false,
			});

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						deployment: null,
						changed: false,
					});
				}
			}, 1000);

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						message: '',
					});
				}
			}, 3000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onAddTag = (): void => {
		let deployment: PodTypes.Deployment;

		if (!this.state.addTag) {
			return;
		}

		if (this.state.changed) {
			deployment = {
				...this.state.deployment,
			};
		} else {
			deployment = {
				...this.props.deployment,
			};
		}

		let tags = [
			...(deployment.tags || []),
		];

		if (tags.indexOf(this.state.addTag) === -1) {
			tags.push(this.state.addTag);
		}

		tags.sort();
		deployment.tags = tags;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addTag: '',
			deployment: deployment,
		});
	}

	onRemoveTag = (tag: string): void => {
		let deployment: PodTypes.Deployment;

		if (this.state.changed) {
			deployment = {
				...this.state.deployment,
			};
		} else {
			deployment = {
				...this.props.deployment,
			};
		}

		let tags = [
			...(deployment.tags || []),
		];

		let i = tags.indexOf(tag);
		if (i === -1) {
			return;
		}

		tags.splice(i, 1);
		deployment.tags = tags;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addTag: '',
			deployment: deployment,
		});
	}

	render(): JSX.Element {
		if (!this.props.open) {
			return <div></div>
		}

		let deployment = this.state.deployment || this.props.deployment

		let tags: JSX.Element[] = [];
		for (let tag of (deployment.tags || [])) {
			tags.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.role}
					key={tag}
				>
					{tag}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveTag(tag);
						}}
					/>
				</div>,
			);
		}

		return <div style={css.settings}>
			<div className="layout vertical wrap">
				<label className="bp5-label">
					Tags
					<Help
						title="Tags"
						content="Deployment tags."
					/>
					<div>
						{tags}
					</div>
				</label>
				<PageInputButton
					disabled={this.state.disabled}
					buttonClass="bp5-intent-success bp5-icon-add"
					label="Add"
					type="text"
					placeholder="Add role"
					value={this.state.addTag}
					onChange={(val): void => {
						this.setState({
							...this.state,
							addTag: val,
						});
					}}
					onSubmit={this.onAddTag}
				/>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.deployment && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						deployment: null,
					});
				}}
				onSave={this.onSave}
			/>
		</div>
	}
}
