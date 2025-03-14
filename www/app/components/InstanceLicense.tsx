/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as License from '../License';
import * as Theme from '../Theme';

interface Props {
	open?: boolean;
	onClose?: () => void;
}

const css = {
	dialog: {
		height: '500px',
		width: '90%',
		maxWidth: '700px',
	} as React.CSSProperties,
	textarea: {
		resize: 'none',
		fontSize: Theme.monospaceSize,
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
};

const license = `ORACLE LINUX LICENSE AGREEMENT

“We,” “us,” “our” and “Oracle” refers to Oracle America, Inc.  “You” and “your” refers to the
individual or entity that has acquired the Oracle Linux programs.  “Oracle Linux programs”
refers to the Linux software product which you have acquired.  “License” refers to your right to
use the Oracle Linux programs under the terms of this Agreement and the licenses referenced
herein.  This Agreement is governed by the substantive and procedural laws of the United States
and the State of California and you and Oracle agree to submit to the exclusive jurisdiction of,
and venue in, the courts of San Francisco or Santa Clara counties in California in any dispute
arising out of or relating to this Agreement.

We are willing to provide a copy of the Oracle Linux programs to you only upon the condition
that you accept all of the terms contained in this Agreement.  Read the terms carefully and
indicate your acceptance by either selecting the “Accept” button at the bottom of the page to
confirm your acceptance, if you are downloading the Oracle Linux programs, or continuing to
install the Oracle Linux programs, if you have received this Agreement during the installation
process.  If you are not willing to be bound by these terms, select the “Do Not Accept” button or
discontinue the installation process.

1. Grant of Licenses to the Oracle Linux programs. Subject to the terms of this Agreement,
Oracle grants to you a license to the Oracle Linux programs under the GNU General Public
License version 2.0. The Oracle Linux programs contain many components developed by Oracle
and various third parties. The license for each component is located in the licensing
documentation and/or in the component's source code.  In addition, a list of components may be
delivered with the Oracle Linux programs and the Additional Oracle Linux programs (as defined
below) or accessed online at http://oss.oracle.com/linux/legal/oracle-list.html.  The source code
for the Oracle Linux Programs and the Additional Oracle Linux programs can be found and
accessed online at https://oss.oracle.com/sources/.  This agreement does not limit, supersede or
modify your rights under the license associated with any separately licensed individual
component.

2. Licenses to Additional Oracle Linux programs.  Certain third-party technology (collectively
the “Additional Oracle Linux programs”) may be included on the same medium or as part of the
download of Oracle Linux programs you receive, but is not part of the Oracle Linux programs.
Each Additional Oracle Linux program is licensed solely under the terms of the Mozilla Public
License, Apache License, Common Public License, GNU Lesser General Public License,
Netscape Public License or similar license that is included with the relevant Additional Oracle
Linux program.

3. Ownership. The Oracle Linux programs and their components and the Additional Oracle
Linux programs are owned by Oracle or its licensors.  Subject to the licenses granted and/or
referenced herein, title to the Oracle Linux programs and their components and the Additional
Oracle Linux programs remains with Oracle and/or its licensors.

4. Trademark License. You are permitted to distribute unmodified Oracle Linux programs or
unmodified Additional Oracle Linux programs without removing the trademark(s) owned by
Oracle or its affiliates that are included in the unmodified Oracle Linux programs or unmodified
Additional Oracle Linux programs (the “Oracle Linux trademarks”). You may only distribute
modified Oracle Linux programs or modified Additional Oracle Linux programs if you remove
relevant images containing the Oracle Linux trademarks. Certain files, identified in
http://oss.oracle.com/linux/legal/oracle-list.html, include such trademarks. Do not delete these
files, as deletion may corrupt the Oracle Linux programs or Additional Oracle Linux programs.
You are not granted any other rights to Oracle Linux trademarks, and you acknowledge that you
shall not gain any proprietary interest in the Oracle Linux trademarks. All goodwill arising out of
use of the Oracle Linux trademarks shall inure to the benefit of Oracle or its affiliates. You may
not use any trademarks owned by Oracle or its affiliates (including “ORACLE”) or potentially
confusing variations (such as, “ORA”) as a part of your logo(s), product name(s), service
name(s), company name, or domain name(s) even if such products, services or domains include,
or are related to, the Oracle Linux programs or Additional Oracle Linux programs.

5. Limited Warranty. THE ORACLE LINUX PROGRAMS AND ADDITIONAL ORACLE
LINUX PROGRAMS ARE PROVIDED "AS IS" WITHOUT WARRANTY OF ANY KIND.
WE FURTHER DISCLAIM ALL WARRANTIES, EXPRESS AND IMPLIED, INCLUDING
WITHOUT LIMITATION, ANY IMPLIED WARRANTIES OF MERCHANTABILITY OR
FITNESS FOR A PARTICULAR PURPOSE.

6. Limitation of Liability. IN NO EVENT SHALL WE BE LIABLE FOR ANY INDIRECT,
INCIDENTAL, SPECIAL, PUNITIVE OR CONSEQUENTIAL DAMAGES, OR DAMAGES
FOR LOSS OF PROFITS, REVENUE, DATA OR DATA USE, INCURRED BY YOU OR
ANY THIRD PARTY, WHETHER IN AN ACTION IN CONTRACT OR TORT, EVEN IF WE
HAVE BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.  OUR ENTIRE
LIABILITY FOR DAMAGES HEREUNDER SHALL IN NO EVENT EXCEED ONE
HUNDRED DOLLARS (U.S.).

7. No Technical Support.  Our technical support organization will not provide technical support,
phone support, or updates to you for the materials licensed under this Agreement.  Technical
support, if available, may be acquired from Oracle or its affiliates under a separate agreement.

8. Relationship Between the Parties. The relationship between you and us is that of
licensee/licensor.  Neither party will represent that it has any authority to assume or create any
obligation, express or implied, on behalf of the other party, nor to represent the other party as
agent, employee, franchisee, or in any other capacity.  Nothing in this Agreement shall be
construed to limit either party's right to independently develop or distribute programs that are
functionally similar to the other party’s products, so long as proprietary information of the other
party is not included in such programs.

9. Entire Agreement.  You agree that this Agreement is the complete Agreement for the Oracle
Linux programs and the Additional Oracle Linux programs, and this Agreement supersedes all
prior or contemporaneous Agreements or representations.  If any term of this Agreement is found
to be invalid or unenforceable, the remaining provisions will remain effective. Neither the
Uniform Computer Information Transactions Act nor the United Nations Convention on the
International Sale of Goods applies to this agreement.

You can find a copy of the GNU General Public License version 2.0 in the “copying” or
“license” file included with the Oracle Linux programs or here:
http://oss.oracle.com/licenses/GPL-2.

OFFER TO PROVIDE SOURCE CODE

For software that you receive from Oracle in binary form that is licensed under an open source
license that gives you the right to receive the source code for that binary, you can obtain a copy
of the applicable source code from https://oss.oracle.com/sources/ or
http://www.oracle.com/goto/opensourcecode.  Alternatively, if the source code for the
technology was not provided to you with the binary, you can also receive a copy of the source
code on physical media by submitting a written request to:

Oracle America, Inc.
Attn: Associate General Counsel
Development and Engineering Legal
500 Oracle Parkway, 10th Floor
Redwood Shores, CA 94065

Or, you may send an email to Oracle using the form linked from
http://www.oracle.com/goto/opensourcecode.  Your written or emailed request should include:
•	The name of the component or binary file(s) for which you are requesting the source code
•	The name and version number of the Oracle product
•	The date you received the Oracle product
•	Your name
•	Your company name (if applicable)
•	Your return mailing address and email
•	A telephone number in the event we need to reach you.
We may charge you a fee to cover the cost of physical media and processing. Your request must
be sent (i) within three (3) years of the date you received the Oracle product that included the
component or binary file(s) that are the subject of your request, or (ii) in the case of code
licensed under the GPL v3, for as long as Oracle offers spare parts or customer support for that
product model or version.

Last updated 24 March 2017`;

export default class InstanceLicense extends React.Component<Props, {}> {
	render(): JSX.Element {
		if (!this.props.open) {
			return <div/>;
		}

		return <div>
			<Blueprint.Dialog
				title="Oracle Linux End-User License Agreement"
				style={css.dialog}
				isOpen={this.props.open}
				usePortal={true}
				portalContainer={document.body}
				onClose={(): void => {
					this.props.onClose();
				}}
			>
				<textarea
					className="bp5-dialog-body bp5-input"
					style={css.textarea}
					autoCapitalize="off"
					spellCheck={false}
					readOnly={true}
					value={license}
				/>
				<div className="bp5-dialog-footer">
					<div className="bp5-dialog-footer-actions">
						<button
							className="bp5-button bp5-intent-danger"
							type="button"
							onClick={(): void => {
								this.props.onClose();
							}}
						>Decline</button>
						<button
							className="bp5-button bp5-intent-success"
							type="button"
							onClick={(): void => {
								License.setOracle(true);
								License.save();
								this.props.onClose();
							}}
						>Accept</button>
					</div>
				</div>
			</Blueprint.Dialog>
		</div>;
	}
}
