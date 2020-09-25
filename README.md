# GoIoTWorkInterruptTrackerLambda
Simple Lambda function to track work context switches to help understand impact to productivity

## Background
One thing I've learned through my time in the professional world is that there are many interruptions to the work cycle during the day. At one point, I wanted to quantify just how bad the context switching was assuming a re-ramp time of 15-30 minutes. Also, tracking time diverted for things like escalations. The problem was that I needed to be able to explain "where the time went" when my metrics were lower than other people. The cause for this was actually because I was a central escalation point for people even at the same job level as myself and I would be interrupted from work or have to work an escalation from a colleague, but would end up receiving no credit for time spent because the mechanism for logging multiple engineer inputs was suboptimal.

## Problem
I needed to create something (but often lacked time/energy) to track these context switches. To create a simple solution (with a more elegant one to come in future revisions), I want to press a button to represent an externally imposed context switch (such as a manager interruption, external team request, escalation, etc.) as a knowledge experiment to understand how much productivity is lost due to various external triggers to context switch during the middle of the day. I want to start by just tracking interruptions, but over time, I want to track activation/deactivation and/or time deltas to track total time lost due to inefficiencies in processes.

## Solution
1. Create a DDB table with a very simple key:
   * device_id+timestamp would be a unique key for the purpose of just tracking number of interruptions in say a week or a month (initial implementation)
   * device_id alone, which would provide the ability to update state through a single attribute (ie. what is the current state of the button: {0,1}) (implement in v1.1)
1. Create a lambda function that triggers an update to ddb for the above value(s)
1. Create a cloudwatch dashboard that can help aggregate data for the last X per time period
1. IAM role for lambda function
1. IAM role for IOT button (could be implemented virtually or through a mobile app as well)

To use this, just compile and load the lambda function to your account and create the rest of the resources to point the lambda function towards.

## How to use
Create resources described above. After creating the resources, attach the iot button to your AWS account and just press the button to test.
