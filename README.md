# Code Walkthrough
Going through the
[Overall solution Design](https://docs.google.com/document/d/1DOT2DHO7pjzqWfw1ynRabP8Ncl8HzdTW4FN9bKCyne8/edit?usp=sharing) is a pre-requisite to understand the implementation. 

## Directory Structure

- **handler** : implementation of endpoints 
  - As a domenstration of this principle we have only shown it for HandleCartEvent endpoint
  - For other endpoints we are directly calling controller to reduce boiler plate code.
- **controller** : contains the main business logic.
  - cart_manager.go - contains the cart controller. It supports CRUD operations on Card and publish events to a Queue [Go Buffered channel](https://notes.kodekloud.com/docs/Advanced-Golang/Concurrency/Buffered-Channels/page)
  - notification_service_controller.go - contains the core functionality for a) handling cart update events b) managing schedules and leases c) providing endpoints for task polling and accepting results. 
  - customer_notifier.go - A placeholder implementation for external communication such as email and push. Currently only logs the contentId and Idempotency key
  - experiment.go - A very basic implementation of experimenation. Only user based experiment supported.
- **entity** : contains all the entities for the 
- **repository** : An in memory storage layer supporting following:
  - dummy_database.go : A generic key value storage implementation. Created so that cart and notification repositories can reuse instead of each rolling out their own. 
  - cart_repository.go - manages storage for carts.
  - notification_repository.go - Contains the notifications status. 
  - redis_sorted_set.go - implementation of Redis sorted set which always maintain elements sorted in ascending order by timestamp.
- **consumer** : consumer subscribing to cart updates event & delegates the business logic to notification handler or controller layer
- **worker** : A worker implementation which polls notification service for new tasks and sends the notification.

## Component Boundaries

- We have decoupled Cart management and notification to provide clear separation of concerns and fault isolation. 
- Notification service is the brain of this system and manages the sole integartion with configuration and experimentation. 
- In an ideal world Cart, notification, consumer and worker would be there own separate applications co-ordinating via APIs or Message Queues.

## A sample Flow

1. A cart is created or updates.
2. A message is published to the Queue i.e. buffered channel
3. Consumer reads from the Queue
4. Consumer calls the notification service to handle the event
5. Notification service creates the schedule for the cart for future execution
6. Wait time expires for the cart and it can be considered abandoned.
7. Worker calls notificationService::GetTask endpoint for tasks.
8. The task is returned and a lease is created for durable execution in case of worker crashes or failures.
9. Worker send the notification via customerNotifier.
10. Worker publish the result to NotificationService.
11. NotificationService record the result and create subsequent schedule. Schedule is defined in abandoned_cart_notification_config.go & chosen based on how many notifications are already sent.
12. NotificationService delete the lease.

Special handling is needed for various scenarios such as:

1. Cart is updated while the notification was already scheduled. Existing schedules needs to reset to align with new state.
2. Out of order events.
3. Gurantee about eventual consistency given we are using polyglot persistence.
4. We have assumed a non zero & significant kafka lag can happen. So we can't rely on events alone to ensure correctness, checks and balances needs to be added based on current state of the system.

## Test Cases

We have written component test cases to test the feature end to end. Below are the key functionalties verified:

1. Customer notification are sent for Abandoned cart. Mutliple notifications are sent with increasing duration.
2. Customer notification aren't sent for Cart which completed before they were abandoned.
3. Customer notifications aren't sent if the user is not in the treatment group. We also handle feature dial down & rollbacks.

