
# Simulator sqlite
This is an alternative simulator implementation, using an event sourcing approach. It writes all incoming requests into a sqlite db. This allows for more flexible constraint definitions. As the upsides for simple constraints on rps are much simpler to implement with a single counter, the sqlite simulator was not used in this thesis.
