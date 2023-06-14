## Munchkin
### A Small Value-Key Store

Munchkin is a small proof-of-concept "value-key store" server. Whereas developers are very familiar with key-value 
stores and their uses, a V-K store inverts the paradigm; rather than constructing a key and asking for the data 
mapped to it, with Munchkin you give it arbitrary JSON data and ask if any keys are associated with it. The core use-cases 
of V-K stores include event dispatching, such as with AWS' EventBridge; security policies and similar document-matching; 
or log/trace routing.

### Quamina
The hard work of Munchkin is owed to Tim Bray (of OpenText and XML fame), whose `Quamia` library (`https://github.com/timbray/quamina`) 
provides the primary in-memory data store and processing engine of Munchkin. Quamina, given a key and a well-formed JSON 
fragment, generates an FSA that acts as a regular expression for the provided JSON, letting you pass arbitrary JSON into 
the server and see if any of the FSAs are able to match it.

`Quamina` is very fast, at the expense of being *extremely* memory-hungry (approximately 60KB of RAM per pattern), 
so be aware that pattern sets reaching into the 100K+ territory will devour a great deal of RAM. However, pattern-matching 
takes about 5-10ms per 10K patterns , so it's more than fast enough for most hot-path lookups in your architecture, 
though if you get into massive cloud-scale numbers you may have to creatively shard your lookups across multiple clusters.

### HTTP Endpoints
There are three HTTP servers exposed by Munchkin, for pattern matching, admin of keys and patterns,
and (not currently supported) for Serf/Raft communications. All of these can be controlled by application flags. We expose different
ports in order to maximize flexibility for virtual network configurations.

##### Admin Calls
- `POST /api/admin/v1/add?key=...` Send JSON pattern as body. Adds the JSON pattern to the database and associates it with the given key.
- `DELETE /api/admin/v1/delete-by-key?key=...` Deletes all JSON patterns for a given key.
##### Match Calls
- `POST /api/v1/match` Send JSON for matching. Will return any matched keys.


### WAL Files
Munchkin has a simple recoverability architecture, with key-pattern adds and deletes written to logfiles that can 
then be loaded into the server at startup. This, like much of Munchkin, is a de minimis implementation, though it 
works effectively enough. WAL files are neither written nor loaded by default.


### TODO
- Add proper logging and observability
- Add credentials management and API
- Add Serf and Raft clustering
