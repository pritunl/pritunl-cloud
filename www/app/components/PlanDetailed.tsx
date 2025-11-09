/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PlanTypes from '../types/PlanTypes';
import * as PlanActions from '../actions/PlanActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import OrganizationsStore from "../stores/OrganizationsStore";
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import PageSelect from "./PageSelect";
import PageTextArea from "./PageTextArea";
import PlanStatement from "./PlanStatement";
import PlanEditor from "./PlanEditor";
import * as Constants from "../Constants";
import * as SecretTypes from "../types/SecretTypes";
import Help from "./Help";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	secrets: SecretTypes.SecretsRo;
	plan: PlanTypes.PlanRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	plan: PlanTypes.Plan;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	wideGroup: {
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class PlanDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			plan: null,
		};
	}

	set(name: string, val: any): void {
		let plan: any;

		if (this.state.changed) {
			plan = {
				...this.state.plan,
			};
		} else {
			plan = {
				...this.props.plan,
			};
		}

		plan[name] = val;

		this.setState({
			...this.state,
			changed: true,
			plan: plan,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PlanActions.commit(this.state.plan).then((): void => {
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
						plan: null,
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

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PlanActions.remove(this.props.plan.id).then((): void => {
			this.setState({
				...this.state,
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
		let plan: PlanTypes.Plan = this.state.plan ||
			this.props.plan;

		let hasOrganizations = false
		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				hasOrganizations = true
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

		if (!hasOrganizations) {
			organizationsSelect.push(
				<option key="null" value="">No Organizations</option>);
		}

		return <td
			className="bp5-cell"
			colSpan={2}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('tab-close') !== -1) {
								this.props.onClose();
							}
						}}
					>
						<div>
							<label
								className="bp5-control bp5-checkbox"
								style={css.select}
							>
								<input
									type="checkbox"
									checked={this.props.selected}
									onChange={(evt): void => {
									}}
									onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
								/>
								<span className="bp5-control-indicator"/>
							</label>
						</div>
						<div className="flex tab-close"/>
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete Plan"
							confirmMsg="Permanently delete this plan"
							confirmInput={true}
							items={[plan.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Plan name."
						type="text"
						placeholder="Enter name"
						value={plan.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Plan comment."
						placeholder="Plan comment"
						rows={3}
						value={plan.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.plan.id || 'Unknown',
							},
						]}
					/>
					<PageSelect
						disabled={this.state.disabled}
						hidden={Constants.user}
						label="Organization"
						help="Organization for plan."
						value={plan.organization}
						onChange={(val): void => {
							this.set('organization', val);
						}}
					>
						{organizationsSelect}
					</PageSelect>
				</div>
			</div>
			<div className="layout horizontal wrap">
				<div className="flex" style={css.wideGroup}>
					<label style={css.itemsLabel}>
						Statements
						<Help
							title="Statements"
							content="Plan statements."
						/>
					</label>
					<PlanEditor
						disabled={this.state.disabled}
						statements={plan.statements}
						onChange={(statements) => {
							let plan: PlanTypes.Plan;

							if (this.state.changed) {
								plan = {
									...this.state.plan,
								};
							} else {
								plan = {
									...this.props.plan,
								};
							}

							plan.statements = statements;

							this.setState({
								...this.state,
								changed: true,
								message: '',
								plan: plan,
							});

						}}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.plan && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						plan: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
