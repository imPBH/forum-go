# Forum

Forum is a simple discussion forum with a simple interface. Users can create topics, give them categories, reply to
topics, and upvote or downvote topics.

## Backend

We use Golang as our backend language. The backend is linked to a Sqlite database in order to store the data and for the
users' authentication. The database is useful for storing the users, topics, replies, and votes.

## Frontend

We use HTML, CSS, and JavaScript to create the frontend. The frontend is linked to the backend in order to interact with
the database.

## Authentication

Once the user logs in, they are given a UUID token in a session cookie. This token is used to authenticate the user. The
token is stored in the database and is used to authenticate the user. When an user logs out, the token is deleted from
the database. When a user is registering, we store is username and hashed password with bcrypt in the database.

## Communication

| Connected | Create post | Add a comment | Create reply | View topic | View comments |
|-----------|-------------|---------------|--------------|------------|---------------|
| ❌         | ❌           | ❌             | ❌            | ✅          | ✅             |
| ✅         | ✅           | ✅             | ✅            | ✅          | ✅             |

## Like and dislike

| Connected | Vote |
|-----------|------|
| ❌         | ❌    |
| ✅         | ✅    |

## Filter posts

| Connected | By categories | Created Post | Liked Posts |
|-----------|---------------|--------------|-------------|
| ❌         | ✅             | ❌            | ❌           |
| ✅         | ✅             | ✅            | ✅           |

## Docker
We use Docker to run the application, we create a Dockerfile in the root directory of the repository.

To create the Docker image, we run the following command:
```bash
docker build -t forum-go .
```
To run the application, we use the following command:
```bash
docker run -p 8000:8000 -it forum-go
```