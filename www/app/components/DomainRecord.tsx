/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DomainTypes from '../types/DomainTypes';

interface Props {
	record: DomainTypes.Record;
	onChange: (record: DomainTypes.Record) => void;
	onRemove: () => void;
}

const css = {
	group: {
		width: '100%',
		marginTop: '5px',
	} as React.CSSProperties,
	type: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	domain: {
		width: '100%',
		borderRadius: '0 3px 3px 0',
	} as React.CSSProperties,
	domainBox: {
		flex: '1',
	} as React.CSSProperties,
};

export default class DomainRecord extends React.Component<Props, {}> {
	clone(): DomainTypes.Record {
		return {
			...this.props.record,
		};
	}

	render(): JSX.Element {
		let record = this.props.record;

		return <div className="bp5-control-group" style={css.group}>
			<div className="bp5-select" style={css.type}>
				<select
					value={record.type}
					onChange={(evt): void => {
						let state = this.clone();
						state.type = evt.target.value;
						if (!state.operation) {
							state.operation = "update"
						}
						this.props.onChange(state);
					}}
				>
					<option value="A">A</option>
					<option value="AAAA">AAAA</option>
					<option value="CNAME">CNAME</option>
				</select>
			</div>
			<div style={css.domainBox}>
				<input
					className="bp5-input"
					style={css.domain}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="Sub Domain"
					value={record.sub_domain || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.sub_domain = evt.target.value;
						if (!state.operation) {
							state.operation = "update"
						}
						this.props.onChange(state);
					}}
				/>
			</div>
			<div style={css.domainBox}>
				<input
					className="bp5-input"
					style={css.domain}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="IP Address"
					value={record.value || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.value = evt.target.value;
						if (!state.operation) {
							state.operation = "update"
						}
						this.props.onChange(state);
					}}
				/>
			</div>
			<button
				className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
				onClick={(): void => {
					this.props.onRemove();
				}}
			/>
		</div>;
	}
}
