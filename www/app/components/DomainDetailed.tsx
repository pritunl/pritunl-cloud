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

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
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
		backgroundColor: 'rgba(0, 0, 0, 0.13)',
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
		let org = OrganizationsStore.organization(this.props.domain.organization);

		return <td
			className="bp3-cell"
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
                className="bp3-control bp3-checkbox"
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
                <span className="bp3-control-indicator"/>
              </label>
            </div>
						<div className="flex tab-close"/>
						<ConfirmButton
							className="bp3-minimal bp3-intent-danger bp3-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp3-intent-danger"
							dialogClassName="bp3-intent-danger bp3-icon-delete"
							dialogLabel="Delete Domain"
							confirmMsg="Permanently delete this domain"
							confirmInput={true}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Domain"
						help="Domain name."
						type="text"
						placeholder="Enter domain"
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
					<PageSelect
						label="Type"
						help="Domain type."
						value={domain.type}
						onChange={(val): void => {
							this.set('type', val);
						}}
					>
						<option value="">Select Type</option>
						<option value="route_53">AWS Route53</option>
					</PageSelect>
					<PageInput
						hidden={domain.type !== 'route_53'}
						label="AWS Access Key ID"
						help="AWS access key ID."
						type="text"
						placeholder="Enter access key ID"
						value={domain.aws_id}
						onChange={(val): void => {
							this.set('aws_id', val);
						}}
					/>
					<PageInput
						hidden={domain.type !== 'route_53'}
						label="AWS Secret Access Key"
						help="AWS secret access key."
						type="text"
						placeholder="Enter secret access key"
						value={domain.aws_secret}
						onChange={(val): void => {
							this.set('aws_secret', val);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.domain.id || 'Unknown',
							},
							{
								label: 'Organization',
								value: org ? org.name : this.props.domain.organization,
							},
						]}
					/>
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
