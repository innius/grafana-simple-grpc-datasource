# 3. move pagination to plugin backend

- Status: proposed
- Deciders: Wolf Fr, Christaan P. and Ron vd W
- Date: 2024-10-15

Technical Story:

The front-end pagination logic exhibited suboptimal performance due to an
excessive memory footprint. The paginated data chunks were concatenated and
emitted to the query observable. While it may appear that these chunks are being
appended to the stream, the reality is that all values are emitted multiple
times. Specifically, the first chunk is emitted, followed by the first and
second chunks, and this pattern continues. For a historical query spanning a
duration of six hours, this results in 75,600 emitted values, rather than the
expected total of 21,600 data points. This inefficiency leads to significant
strain on the system as all these values traverse through it. This Architectural
Decision Record (ADR) explores alternative solutions to address this issue
effectively.

## Context and Problem Statement

Refine the pagination logic to ensure robust query execution on the dashboard,
thereby preventing application crashes. This enhancement will optimize
performance and improve user experience by allowing seamless interaction with
large datasets.

## Decision Drivers

- The implementation must ensure that existing customer dashboards remain fully
  operational and unaffected.
- Additionally, it is imperative that the user experience is preserved and not
  diminished in any way.
- Furthermore, the solution should avoid introducing any unnecessary complexity
  to the frontend architecture.

## Considered Options

- add client side cache
- remove streaming behavior for query paginator
- move pagination to plugin backend

## Decision Outcome

Selected Option: Transitioning pagination to the plugin backend is the preferred
option, as it substantially simplifies frontend complexity while ensuring
efficient handling of pagination processes. Furthermore, this change alleviates
a critical limitation of the existing system, specifically the inability to
utilize expressions within paginated queries. This enhancement not only
streamlines the user experience but also optimizes overall system performance.

### Positive Consequences <!-- optional -->

- improves overall query performance.
- eliminates the need of frontend pagination logic.
- supports expressions for paginated queries.

### Negative Consequences <!-- optional -->

- The user experience associated with slow queries, particularly those that
  execute over extended periods, is notably less responsive. Presently, users
  observe incremental data being appended to the timeseries in real-time. In
  contrast, under the proposed changes, the panel remains devoid of content
  until the query execution is fully completed. This shift may lead to a
  perception of latency and could adversely affect user engagement and
  satisfaction.

## Pros and Cons of the Options <!-- optional -->

### add client side cache

Implement client-side caching mechanisms within the frontend architecture to
optimize data retrieval processes. This enhancement will involve caching query
results, thereby reducing redundant network requests. Subsequent queries will be
streamlined to only request data from the head of the selected range, ensuring
efficient use of resources and improved application performance.

- Good, because it reduces the number or network requests significantly
- Good, because it improves the query performance significantly
- Bad, because it significantly escalates the complexity of the frontend logic.
  This is a critical consideration, particularly given that TypeScript is not
  our primary development environment. Consequently, the process of debugging
  and implementing new features is considerably more time-consuming, which could
  impact our overall project timelines and resource allocation.

### remove streaming behavior for query paginator

This approach eliminates the streaming behavior inherent in the query paginator
logic. Rather than emitting values incrementally after each iteration, all
values are emitted collectively upon the completion of the iteration. This
modification enhances the efficiency of data handling and ensures that the
entire dataset is processed before any output is generated.

- Good, because it requires only a small change of the current codebase.
- Good, because it reduces the memory footprint significantly.
- Good, because it can easily be extended with client caching capabilities.
- Bad, because it still requires complex frontend pagination logic.

### move pagination to plugin backend

This approach effectively delegates the pagination logic to the plugin backend,
which is implemented in Go. The backend is responsible for iterating through the
pages returned by the API and consolidating the entire dataset before
transmitting it to the frontend in a single response. Consequently, this
eliminates the need for any additional pagination logic on the frontend. In
fact, this design allows the frontend data source to revert to the standard
Grafana data source implementation, thereby streamlining the overall
architecture and enhancing maintainability.

- Good, it simplifies the frontend architecture
- Good, it handles pagination efficiency
- Good, because it can be done in our default developer environment
- Good, because it enables expressions for queries which spans multiple pages.
- Bad, because it adversely affect the user experience, particularly when
  querying large datasets. Under these circumstances, there is a significant
  delay in the feedback provided to the user from the server. Presently, users
  observe the incremental addition of pages to the time series in real-time,
  which can lead to confusion and frustration.
