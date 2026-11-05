# Subscribe through AWS Marketplace

> Set up Orion pay-as-you-go billing in AWS Marketplace


With pay-as-you-go billing, you can subscribe to Orion and get billed through your [AWS account](https://aws.amazon.com/).
Billing is based on an hourly rate, per the number of requested CPU cores for each cluster connected to your Orion account. There is no upfront or minimum cost.

See the [AWS Marketplace - Buyers Guide][awsbuyerguide]
for details on the AWS Marketplace and the [Orion product page][orionmarketplaceproduct] for Orion pricing.

## Subscribing to Orion

1. Navigate to the [Orion product page][orionmarketplaceproduct] in the AWS Marketplace.
2. Click **View purchase options**.
3. Review the terms and pricing; then, click **Subscribe**.

    <div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Billing an existing Orion account</div>


    If you already have a Orion account that does **not** have pay-as-you-go billing, your last step is to  contact [Orion Support][support] and provide your AWS account ID. Orion Support will convert your existing account to a subscription.
    </div>


4. Click the **Set up your account** link, which takes you to the Orion sign up page.
5. Sign up for a new Orion account by providing a username and password.
   At this time, social logins (social SSO) are not supported when creating Orion subscriptions.


6. Confirm your email address via the link sent to you in the welcome email.
7. After logging in, confirm that the account menu shows _This account is billed through the AWS Marketplace_:

    ![Image showing the AWS Account menu with text indicating "This account is billed through the AWS Marketplace"](/img/docs/ol-billed-through-aws-marketplace.png)

8. Continue with the [Orion installation][install-guide].

After 48 hours, you should see usage data for the Orion product in your AWS Billing and Cost Management.
You can use AWS [Cost Explorer][awscostexplorer] to analyze the hourly charges.
In the **Marketplace charges by service** section under [**Bills**][awsbills], you can see the total charges for the current billing period.

## Canceling your subscription

To cancel your subscription, navigate to the AWS Marketplace subscriptions page in the AWS console and click **Cancel subscription** in the **Actions** menu. It takes about one hour to close your Orion account, at which time you will no longer be billed through AWS Marketplace.

For a step-by-step-guide, see [Canceling your SaaS subscription][awsbuyercancelation] in the AWS documentation.

Lastly, [uninstall] the Orion Agent from your clusters.

## Configuring cost allocation tags for per-cluster reporting

By default, AWS Marketplace exposes the total cost billed to your account only. To enable per-cluster reporting, configure a cost allocation tag for `aws:marketplace:isv:cluster`.

**NOTE:** It can take up to 48 hours before you see per-cluster billing in AWS Cost Explorer. This is because it typically takes 24-28 hours for AWS Cost Explorer to process and show Orion usage reports. AWS Cost Explorer will apply the new tag to Orion usage reports that it has not yet processed, as well as future usage reports.

For details, see the AWS [Vendor-metered tags](https://docs.aws.amazon.com/marketplace/latest/buyerguide/cost-allocation-tagging.html#vendor-metered-tags) guide.


[install-guide]: /docs/installation/install/#install-orion
[orionmarketplaceproduct]: https://aws.amazon.com/marketplace/pp/prodview-pcp2mkraouwbk?trk=8d276e92-b310-40ce-908f-23a198ca7ffc&sc_channel=el&source=orion
[awsbills]: https://console.aws.amazon.com/billing/home#/bills
[awscostexplorer]: https://console.aws.amazon.com/costmanagement/home#/cost-explorer
[awsbuyerguide]: https://docs.aws.amazon.com/marketplace/latest/buyerguide/saas-subscriptions.html
[awsbuyercancelation]: https://docs.aws.amazon.com/marketplace/latest/buyerguide/cancel-subscription.html#cancel-saas-subscription
[uninstall]: /docs/installation/uninstall/
[support]: https://www.orioncorp.io/company/contact-us/
