/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DomainTypes from '../types/DomainTypes';
import * as DomainActions from '../actions/DomainActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import OrganizationsStore from "../stores/OrganizationsStore";
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import PageSelect from "./PageSelect";
import PageTextArea from "./PageTextArea";
import DomainRecord from "./DomainRecord";
import * as Constants from "../Constants";
import * as SecretTypes from "../types/SecretTypes";
import Help from "./Help";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	secrets: SecretTypes.SecretsRo;
	domain: DomainTypes.DomainRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	domain: DomainTypes.Domain;
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

export default class DomainDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			domain: null,
		};
	}

	set(name: string, val: any): void {
		let domain: any;

		if (this.state.changed) {
			domain = {
				...this.state.domain,
			};
		} else {
			domain = {
				...this.props.domain,
			};
		}

		domain[name] = val;

		this.setState({
			...this.state,
			changed: true,
			domain: domain,
		});
	}

	onAddRecord = (): void => {
		let domain: DomainTypes.Domain;

		if (this.state.changed) {
			domain = {
				...this.state.domain,
			};
		} else {
			domain = {
				...this.props.domain,
			};
		}

		let records = [
			...(domain.records || []),
			{
				operation: "insert",
			},
		];

		domain.records = records;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			domain: domain,
		});
	}

	onChangeRecord(i: number, state: DomainTypes.Record): void {
		let domain: DomainTypes.Domain;

		if (this.state.changed) {
			domain = {
				...this.state.domain,
			};
		} else {
			domain = {
				...this.props.domain,
			};
		}

		let records = [
			...(domain.records || []),
		];

		state.type = (state.type || "A")

		records[i] = state;

		domain.records = records;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			domain: domain,
		});
	}

	onRemoveRecord(i: number): void {
		let domain: DomainTypes.Domain;

		if (this.state.changed) {
			domain = {
				...this.state.domain,
			};
		} else {
			domain = {
				...this.props.domain,
			};
		}

		let records = [
			...(domain.records || []),
		];

		records[i] = {
			...records[i],
			operation: "delete",
		};

		domain.records = records;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			domain: domain,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DomainActions.commit(this.state.domain).then((): void => {
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
						domain: null,
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
		DomainActions.remove(this.props.domain.id).then((): void => {
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
		let domain: DomainTypes.Domain = this.state.domain ||
			this.props.domain;

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

		let hasSecrets = false;
		let secretsSelect: JSX.Element[] = [];
		if (this.props.secrets.length) {
			secretsSelect.push(<option key="null" value="">Select Secret</option>);

			for (let secret of this.props.secrets) {
				if (Constants.user) {
					if (domain.organization !== OrganizationsStore.current) {
						continue;
					}
				} else {
					if (domain.organization !== secret.organization) {
						continue;
					}
				}

				hasSecrets = true;
				secretsSelect.push(
					<option
						key={secret.id}
						value={secret.id}
					>{secret.name}</option>,
				);
			}
		}

		if (!hasSecrets) {
			secretsSelect = [<option key="null" value="">No Secrets</option>];
		}

		let records: JSX.Element[] = [];
		for (let i = 0; i < domain.records.length; i++) {
			let index = i;

			if (domain.records[index].operation === "delete") {
				continue;
			}

			records.push(
				<DomainRecord
					key={index}
					record={domain.records[index]}
					onChange={(state: DomainTypes.Record): void => {
						this.onChangeRecord(index, state);
					}}
					onRemove={(): void => {
						this.onRemoveRecord(index);
					}}
				/>,
			);
		}

		return <td
			className="bp5-cell"
			colSpan={5}
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
							dialogLabel="Delete Domain"
							confirmMsg="Permanently delete this domain"
							confirmInput={true}
							items={[domain.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Domain name."
						type="text"
						placeholder="Enter name"
						value={domain.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Domain comment."
						placeholder="Domain comment"
						rows={3}
						value={domain.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageInput
						label="Domain"
						help="Root domain name."
						type="text"
						placeholder="Enter domain"
						value={domain.root_domain}
						onChange={(val): void => {
							this.set('root_domain', val);
						}}
					/>
					<label style={css.itemsLabel}>
						Domain Records
						<Help
							title="Domain Records"
							content="Domain DNS records."
						/>
					</label>
					{records}
					<button
						className="bp5-button bp5-intent-success bp5-icon-add"
						style={css.itemsAdd}
						type="button"
						onClick={this.onAddRecord}
					>
						Add Record
					</button>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.domain.id || 'Unknown',
							},
						]}
					/>
					<PageSelect
						disabled={this.state.disabled}
						hidden={Constants.user}
						label="Organization"
						help="Organization for domain."
						value={domain.organization}
						onChange={(val): void => {
							this.set('organization', val);
						}}
					>
						{organizationsSelect}
					</PageSelect>
					<PageSelect
						label="Provider"
						help="Domain provider."
						value={domain.type}
						onChange={(val): void => {
							this.set('type', val);
						}}
					>
						<option value="aws">AWS</option>
						<option value="cloudflare">Cloudflare</option>
						<option value="oracle_cloud">Oracle Cloud</option>
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Provider API Secret"
						help="Secret containing API keys to use for provider."
						value={domain.secret}
						onChange={(val): void => {
							this.set('secret', val);
						}}
					>
						{secretsSelect}
					</PageSelect>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.domain && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						domain: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
