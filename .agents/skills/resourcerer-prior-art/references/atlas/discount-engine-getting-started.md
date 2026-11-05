---
updatedAt: 2026-05-12T14:12:41.000Z
---

# Getting started

Atlas's Price adjustment feature allows you to apply custom pricing to your cloud resources, helping align cost reporting and optimization decisions with your actual cloud costs. This guide walks you through setting up your first price adjustment in the Atlas console.

# Prerequisites

* Atlas account with the necessary permissions within your organization
* At least one connected cluster
* Access to pricing for your cloud resources (e.g., enterprise discount program details)

# Create your first adjustment

1. Navigate to **Optimization** > **Price adjustments** in the Atlas console sidebar
2. Click the **Add adjustment** button in the top right corner

<Image border={false} src="https://files.readme.io/6e1568de9f01415adb8ac71e38890334b0943d7d5aa0a34487d549f11b5d3975-image.png" />

## Basic adjustment configuration

In the drawer that opens, configure the options for the creation of your adjustment.

### Enter a descriptive name for your adjustment entry

This name will help you identify the adjustment in reports and the node list.

<Image border={false} src="https://files.readme.io/bb046abf69a7242b8a07befbeb0b5884b2780de632ca7b1cc71961af510b775c-image.png" />

### Configure resource filters

1. **Select your cloud provider** (AWS, GCP, or Azure)

   <Image border={false} src="https://files.readme.io/9c51b5812c895f0c867f28778838adddf8ea70e022b119caead966b1591a54cd-image.png" />
2. **Define instance constraints**
   * Filter by instance families (e.g., m5) or specify individual instance types (optional)
   * Choose from resource offerings:
     * On-demand
     * Spot

<Image border={false} src="https://files.readme.io/df69d369ba01d694d2788a43b3d0e0214f9c955070134a77db5ed4fe5721231d-image.png" />

## Define adjustment parameters

1. **Select adjustment type**
   * Choose the pricing adjustment you want to apply:
     * **Discount**: Reduce the public list price
     * **Surcharge**: Increase the public list price
     * **Custom pricing**: Replace the public list price with a specific value

2. **Specify adjustment details**
   * Choose what to adjust:
     * CPU
     * Memory
     * Instance (applies to the whole instance)
   * Select value type:
     * **Flat**: Apply a fixed dollar amount ($/hour)
     * **Percentage**: Apply a percentage change
   * Enter the adjustment value

3. **Set active period**
   * Define when the adjustment starts
   * Set an end date (optional)

> 📘 Note
>
> Price adjustments are **not** retroactive and only apply to future costs.

<Image border={false} src="https://files.readme.io/78ede8243b364659afd529e177902108a54c647344986aef14b1d9af9cf09dda-image.png" />

## Apply cluster constraints

1. **Set geographical scope**

   * Select specific regions, choose availability zones, or leave as "All" to apply the adjustment globally.

   <Image border={false} src="https://files.readme.io/9619362feec5d47533f80449b340fc9823bb41cec195c456580b79d15bf4d226-image.png" />

   <br />

2. **Choose target clusters**
   * Select specific clusters to apply the adjustment to.
   * Check "Select to include specific clusters" to limit scope or leave unchecked to apply to all current and future clusters.

3. Click **Save and apply** to activate your adjustment.

<Image border={false} src="https://files.readme.io/96967c2bb936faa4093d41178ab36f50821bbac83e0b97ac332eca1dc7ea86fd-image.png" />

# Verify adjustment application

After creating an adjustment, you can verify if it's applied and active:

1. Navigate to **Node list** in your cluster view
2. Look for the adjustment indicator in the new **Active price adjustments** column within **Pricing**

<Image border={false} src="https://files.readme.io/f40ef9984afdb144b0038758c71206be7c228b6ae16dcde0783059c9dd04a242-image.png" />

Hover over the adjustment icon to see additional information at a glance, such as the number of active adjustments and adjustment details. Moreover, hovering over the hourly resource rates reveals the base price.

> 📘 Note
>
> Allow 5-10 minutes for new adjustments to propagate through the system. Cost reporting data may take up to an hour to reflect new adjustment configurations due to caching.

For additional help or to request access to the Price adjustments feature, contact Atlas support.