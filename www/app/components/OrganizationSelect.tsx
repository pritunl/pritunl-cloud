/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from "@blueprintjs/core"
import * as BpSelect from '@blueprintjs/select';
import * as Icons from '@blueprintjs/icons';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as CompletionActions from '../actions/CompletionActions';
import CompletionStore from '../stores/CompletionStore';

interface Props {
	hidden: boolean;
}

interface State {
	organizations: OrganizationTypes.Organizations;
	organization: string;
}

const css = {
	cardButton: {
	} as React.CSSProperties,
	select: {
		display: "inline",
	} as React.CSSProperties,
};

export default class Organization extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			organizations: [...CompletionStore.organizations],
			organization: null,
		};
	}

	componentDidMount(): void {
		CompletionStore.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			organizations: [...CompletionStore.organizations],
			organization: CompletionStore.userOrganization,
		});
	}

	renderOrg: BpSelect.ItemRenderer<OrganizationTypes.Organization> = (org,
		{handleClick, handleFocus, modifiers, query, index}): JSX.Element => {

		let className = ""
		let selected = false
		if (this.state.organization === org.id) {
			className = "bp5-text-intent-primary bp5-intent-primary"
			selected = true
		} else if (index === 0) {
			className = ""
		}
		return <Blueprint.MenuItem
			key={`org-${org.id}`}
			selected={selected}
			roleStructure="listoption"
			icon={<Icons.People
				className={className}
			/>}
			onFocus={handleFocus}
			onClick={(evt): void => {
				evt.preventDefault()
				evt.stopPropagation()
				handleClick(evt)
			}}
			text={org.name}
			textClassName={className}
		/>
	}

	render(): JSX.Element {
		if (this.props.hidden) {
			return <div/>
		}

		return <BpSelect.Select<OrganizationTypes.OrganizationRo>
			items={this.state.organizations || []}
			itemRenderer={this.renderOrg}
			popoverTargetProps={{
				style: css.select,
			}}
			filterable={false}
			itemListRenderer={({items, itemsParentRef,
					query, renderItem, menuProps}) => {

				const renderedItems = items.map(renderItem).filter(
					item => item != null)
				return <Blueprint.Menu
					role="listbox"
					ulRef={itemsParentRef}
					{...menuProps}
				>
					{renderedItems}
				</Blueprint.Menu>
			}}
			onItemSelect={(org) => {
				CompletionActions.setUserOrganization(org.id);
			}}
		>
			<Blueprint.Button
				style={css.cardButton}
				alignText="left"
				small={true}
				rightIcon={<Icons.CaretDown/>}
			>Organization</Blueprint.Button>
		</BpSelect.Select>
	}
}
