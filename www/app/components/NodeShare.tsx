/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import * as BlockTypes from '../types/BlockTypes';
import PageInputButton from './PageInputButton';

interface Props {
	disabled: boolean;
	share: NodeTypes.Share;
	onChange: (state: NodeTypes.Share) => void;
	onAdd: () => void;
	onRemove: () => void;
}

interface State {
	addRole: string;
}

const css = {
	group: {
		width: '100%',
		maxWidth: '216px',
		marginTop: '5px',
		marginBottom: '15px',
	} as React.CSSProperties,
	sourceGroup: {
		width: '100%',
		maxWidth: '219px',
		marginTop: '5px',
	} as React.CSSProperties,
	control: {
		marginTop: '0px',
	} as React.CSSProperties,
	protocol: {
		marginTop: '0px',
		flex: '0 1 auto',
	} as React.CSSProperties,
	input: {
		width: '100%',
	} as React.CSSProperties,
	inputBox: {
		flex: '1',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	roles: {
		width: '100%',
		maxWidth: '400px',
		marginTop: '5px',
		marginBottom: '15px',
	} as React.CSSProperties,
	other: {
		flex: '0 1 auto',
		width: '52px',
		borderRadius: '0 3px 3px 0',
	} as React.CSSProperties,
};

export default class NodeShare extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			addRole: null,
		};
	}

	clone(): NodeTypes.Share {
		return {
			...this.props.share,
		};
	}

	onAddRole = (): void => {
		let state = this.clone();

		if (!this.state.addRole) {
			return;
		}

		let roles = [
			...(state.roles || []),
		];

		if (roles.indexOf(this.state.addRole) === -1) {
			roles.push(this.state.addRole);
		}

		roles.sort();
		state.roles = roles;

		this.props.onChange(state);
	}

	onRemoveRole = (role: string): void => {
		let state = this.clone();

		let roles = [
			...(state.roles || []),
		];

		let i = roles.indexOf(role);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);
		state.roles = roles;

		this.props.onChange(state);
	}

	render(): JSX.Element {
		let share = this.props.share;

		let roles: JSX.Element[] = [];
		for (let role of (share.roles || [])) {
			roles.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.role}
					key={role}
				>
					{role}
					<button
						className="bp5-tag-remove"
						disabled={this.props.disabled}
						onMouseUp={(): void => {
							this.onRemoveRole(role);
						}}
					/>
				</div>,
			);
		}

		return <div>
			<div style={css.roles}>
				{roles}
			</div>
			<PageInputButton
				disabled={this.props.disabled}
				buttonClass="bp5-intent-success bp5-icon-add"
				label="Add"
				type="text"
				listStyle={true}
				placeholder="Add role"
				value={this.state.addRole}
				onChange={(val): void => {
					this.setState({
						...this.state,
						addRole: val,
					});
				}}
				onSubmit={this.onAddRole}
			/>
			<div className="bp5-control-group" style={css.group}>
				<input
					className="bp5-input"
					style={css.input}
					disabled={this.props.disabled}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="Enter share path"
					value={share.path || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.path = evt.target.value;
						this.props.onChange(state);
					}}
				/>
				<button
					className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
					style={css.control}
					onClick={(): void => {
						this.props.onRemove();
					}}
				/>
				<button
					className="bp5-button bp5-minimal bp5-intent-success bp5-icon-add"
					style={css.control}
					onClick={(): void => {
						this.props.onAdd();
					}}
				/>
			</div>
		</div>;
	}
}
