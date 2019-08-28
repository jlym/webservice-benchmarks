What happens when lots of http requests hit a server?
At what point does the server break?
What are the various resources that get used?
Which finish first?


Variables:
  * number of workers
  * what the handler does
  * type of request - HTTP Get, 

Things to monitor:
  * TCP connections
  * Successes vs failures
  * Request latency
  * Errors


Number of workers vs
  * Success request rate
  * Average latency
  * Number of connections in each state 1 min before the end of the run
  *  All workers at the same rate

Dashboard:

  * Summary
    * Number of workers vs each other thing
  
  * Rates
  * Error message
  * TCP connections over time

Dashboard: