# Schedule

> When Orion generates recommendations and how long they stay valid


Following the _learning period_, Orion generates recommendations at the interval defined by the _schedule_. An optional recommendation _validity period_ can be configured independently from the recommendation schedule.

For more details and examples showing how these settings are related, see the [Schedule][schedule], [Validity period][validity-period], [Learning period][learning-period], and [Preliminary recommendations][prel-recs-concepts] sections in the Concepts topic.

**Recommendation settings:**

- [Recommendation schedule](#schedule)
- [Validity period](#validity-period)

## Schedule {#schedule}

The schedule on which Orion generates new recommendations for a workload.

| Annotation                    | Default value                      |
| ----------------------------- | ---------------------------------- |
| `live.orion.example/schedule` | `"@daily"` (equivalent to `P1D`)   |

### Description {#schedule-description}

Recommendations are generated automatically for workloads on a schedule, defined by a schedule string. The shortest valid recommendation schedule is once an hour.

To prevent the restarting of many workloads at the same time, Orion staggers the generation and applying of recommendations throughout the schedule period (when the schedule is an ISO 8601 Duration string or equivalent schedule macro). Examples:

- `@daily` or `P1D`: recommendations are generated and applied across 24 hours
- `PT12H`: recommendations are generated and applied across 12 hours

If you require more precise timing for when recommendations are applied, use a cron expression as described below.

### Duration expressions

Valid ISO 8601 [Duration][duration] strings can be specified.

#### Examples {#schedule-duration-examples}

- `PT1H` – Once an hour
- `P1D` – Once a day
- `P7D` – Once a week

### Cron expressions

[Cron][cron] strings can be used to designate schedules with more specific requirements, such as day-of-week. Cron expressions specify five fields:

```text
┌───────── minute (0–59)
│ ┌─────── hour (0–23)
│ │ ┌───── day of month (1–31)
│ │ │ ┌─── month (1–12)
│ │ │ │ ┌─ day of week (0–7, where 0 and 7 are both Sunday)
│ │ │ │ │
* * * * *
```

For day-of-week, friendly terms such as `WED` or `THU` are acceptable in place of numbers, and appending `L` selects the last such day of the month (for example, `SATL` is the last Saturday). By default, cron times are interpreted in UTC; use [`schedule.timezone`](#schedule-time-zone) to specify a time zone.

To spread load across workloads, use `H` instead of fixed values when you don't need exact timing.

#### H (hash) syntax

`H` stands for "hash." In any field, `H` tells Orion to choose a value for you, rather than using a fixed value. The value for `H` is derived from a workload's identity, so it is:

- **Stable** – the same workload always resolves `H` to the same value.
- **Distributed** – different workloads resolve the same expression to different values.

For example, `H H * * *` runs each workload once a day, at a different time for each workload.

##### Forms of H

| Form             | Meaning                                                                  |
| ---------------- | ------------------------------------------------------------------------ |
| `H`              | A hashed value across the field's whole range                            |
| `H(start-end)`   | A hashed value within the given range                                    |
| `H/x`            | A hashed starting point, then every `x` units through the field's range  |
| `H(start-end)/x` | A hashed starting point within a range, then every `x` units             |


##### Range that H selects from

For minute, hour, and month, `H` selects from the same range as the field. Two fields differ:

- **Day of month** – `H` selects from 1–28, not 1–31, to ensure the schedule fires every month regardless of length; it never lands on the 29th, 30th, or 31st. A fixed value such as `31` is used as written.
- **Day of week** – `H` selects from 0–6 (Sunday–Saturday). Cron accepts `7` as an alias for Sunday, but `H` only needs one representation of Sunday, so it never produces `7`.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Zero in the minute field is hashed</div>


To prevent load spikes, Orion hashes the minute field when `0` or `H` is provided. For example, `0 2 * * *` runs at a hashed minute during the 2:00am hour, not exactly at 2:00am. Use a non-zero value, such as `5 2 * * *`, to pin to a specific minute. All other fields are used exactly as written unless you specify `H`.
</div>


#### Examples {#schedule-cron-examples}

- `H * * * *` – Once an hour
- `H H(9-17)/2 * * 1-5` – Every 2 hours between 9:00am and 5:59pm, Monday–Friday
- `H H/6 * * *` – Every 6 hours
- `H H(0-2) * * *` – Once a day, between 12:00am and 2:59am
- `30 0 * * SAT,MON` – 12:30am each Saturday and each Monday (fixed time)
- `H H(7-9) * * SATL` – Between 7:00am and 9:59am on the last Saturday of the month

### Schedule macros

Macros are shorthand that can be used as syntactic sugar for common schedules. The available macros are described below and compared with their duration and cron equivalents.

| Macro       | Duration           | Cron             | Description                                    |
| ----------- | ------------------ | ---------------- | ---------------------------------------------- |
| `@never`    | `P0D`              | N/A              | Do not generate recommendations automatically  |
| `@hourly`   | `PT1H`             | `H * * * *`      | Once every hour                                |
| `@daily`    | `P1D`              | `H H * * *`      | Once every day                                 |
| `@weekly`   | `P7D`              | `H H * * H`      | Once every week                                |
| `@midnight` | N/A                | `H H(0-2) * * *` | Every day between 12:00am and 2:59am UTC       |

### Time zone {#schedule-time-zone}

By default, the recommendation schedule is evaluated in UTC. To use a different time zone, set `schedule.timezone` to an [IANA Time Zone][iana-tz] identifier.

| Annotation                             | Default value |
| -------------------------------------- | ------------- |
| `live.orion.example/schedule.timezone` | `"UTC"`       |

The time zone applies to cron expressions, duration expressions, and macros. Daylight saving time transitions for the specified time zone are honored.

#### Examples {#schedule-time-zone-examples}

- `America/New_York` – US Eastern time
- `Europe/London` – UK time
- `Asia/Tokyo` – Japan Standard Time

For example, pairing the schedule `30 0 * * *` with the time zone `America/New_York` generates recommendations at 12:30am US Eastern.

### Key points {#schedule-key-points}

- Each workload can have one and only one schedule string defined for it, of either a duration, cron expression, or schedule macro.
- The schedule time zone defaults to UTC; configure it using [`schedule.timezone`](#schedule-time-zone).
- Exact times are not guaranteed for recommendation generation. For exact cron strings, recommendations are generated within an hour of the time specified.
- To avoid concentrated pod churn when new recommendations are applied, use schedules that spread optimization over a period of time. The `H` capability in cron expressions or using non-specific durations can both help achieve an appropriate scheduling distribution.

## Validity period {#validity-period}

Defines the duration for which a recommendation remains valid.

| Annotation                           | Default value |
| ------------------------------------ | ------------- |
| `live.orion.example/validity-period` | None          |

### Description {#validity-period-description}

Overrides the default recommendation validity period where recommendations remain valid until new recommendations are generated according to the recommendation schedule.

Validity period must be greater than the recommendation schedule frequency for the validity period to take effect. For example, if the recommendation schedule is `@daily`, validity period must be set to at least `P1DT1H`.

If validity period is configured to be equal to or less than the recommendation schedule, validity continues to be based on the recommendation schedule.

### Duration expressions {#validity-period-duration-examples}

Only valid ISO 8601 [Duration][duration] strings can be specified.

#### Examples {#validity-period-examples}

- "`P5D`" – Five days
- "`P14D`" – Two weeks

[schedule]: /docs/concepts/#schedule
[validity-period]: /docs/concepts/#validity-period
[learning-period]: /docs/concepts/#learning-period
[prel-recs-concepts]: /docs/recommendations/#prel-recs-concepts
[auto-deploy]: /docs/configure/settings/apply-settings/#auto-deploy
[cron]: <https://en.wikipedia.org/wiki/Cron>
[duration]: <https://en.wikipedia.org/wiki/ISO_8601#Durations>
[iana-tz]: <https://en.wikipedia.org/wiki/List_of_tz_database_time_zones>
