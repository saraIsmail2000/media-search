# Media Search
The project is a Go exercise focusing on goroutines, caching, Elasticsearch, and cron jobs.

### Overview
The project comprises a job for migrating data from two CSV files, one for movies and another for books, into both a database and an Elasticsearch datastore. 
The migration job utilizes goroutines with buffered channels for concurrent reading and insertion of data from the CSV files.

Additionally, `APIs` are implemented using the Swagger and Echo frameworks, providing functionality for three different endpoints:

### Search API:
This endpoint allows users to search for a book or movie using a search string as a query parameter. 
The API checks if the search query is cached, if not, 
it queries the Elasticsearch client on the indexes of both the books and movies tables. 
The API `returns` a combined list of results containing the:
- ID
- title
- authors
- cast
- writers
- and the `type` of the result (book or movie) 
along with a randomly generated search ID.

`note that for caching we are using redis server`

### Save Search Event API: 
This endpoint is used to save search events in the `search_events` table.
Taking search ID (primary key) and search query in the request body. 
The timestamp column is automatically populated with the current time when the event is inserted into the database.

### Save Click Event API: 
This endpoint saves when a search click event in the table `search_clicks`. 
Taking the search ID, result position, result type (book or movie), and result ID (the actual ID of the book or movie in the database) in the request body.


### Cron Job
In addition to the migration job and APIs, there is a cron job for generating insights into the search process. 
This cron job is scheduled to run every 24 hours and calculates the following insights:

- Top 10 clicked items from search results in the last 24 hours.
- Average click position per day.
- Total searches and clicks per day.
- Click-through rate (CTR) calculation (out of X searches per day, how many clicks there are).

These insights are saved in a JSON file named after the current time of the job execution and stored under the insights folder in the project directory.

## Dependencies
To download all dependencies for the go project, run the following command:
```
go mod download
```

### Docker
I'm using a Docker container for the MySQL instance.

To run the Docker container with Docker Compose, navigate to the directory containing your docker-compose.yml file and run:
```
docker-compose up -d
```

### Elastic Search
I'm using [go-elasticsearch](https://github.com/elastic/go-elasticsearch) package version v7.17

To  install elastic search engine go to [Official Downloads](https://www.elastic.co/downloads/elasticsearch) ,

i installed it as a zip folder and then ran `elasticsearch.bat` under bin folder.

### Usage
To execute the migration job, run:
```
go run main.go migrate
```

To run the server for testing the APIs, use:
```
go run main.go run
```
and go to
http://127.0.0.1:8081/api/v1/swagger/index.html
