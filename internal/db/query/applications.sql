-- name: CreateApplication :one
INSERT INTO applications(name,api_key)
VALUES($1,$2)
RETURNING id, name, api_key, created_at;

-- name: GetApplicationByapikey :one
select * from applications 
where api_key = $1;

-- name: GetApplications :many
select * from applications
order by created_at desc;

-- name: GetApplicationByID :one
select * from applications
where id = $1;

-- name: UpdateApplication :exec
update applications
set name = $1, api_key = $2
where id = $3;

-- name: DeleteApplication :exec
delete from applications
where id = $1;
