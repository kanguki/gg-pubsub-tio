```sh
# set up infrastructure
proj="your-project-name"
unique_key="pubsubtio" # must have at least 4 characters

cd infra
terraform plan -var project_name=$proj -var unique_key=$unique_key
terraform apply -var project_name=$proj -var unique_key=$unique_key --auto-approve
export GOOGLE_APPLICATION_CREDENTIALS=`pwd`/credentials.json

cd ../app
go run . publish -p=$proj -t=$unique_key -s=$unique_key --batch=100 --times=10

# start 3*4=12 consumers 
go run . subscribe -p=$proj -t=$unique_key -s=$unique_key --parallel=4 &
go run . subscribe -p=$proj -t=$unique_key -s=$unique_key --parallel=4 &
go run . subscribe -p=$proj -t=$unique_key -s=$unique_key --parallel=4 &
sleep 10

for job in `jobs -p`
do
    echo $job
    kill $job
done

cd ../infra
unset GOOGLE_APPLICATION_CREDENTIALS
terraform destroy -var project_name=$proj -var unique_key=$unique_key --auto-approve

```

demo running

![1705926571049](image/readme/1705926571049.png)

---

using google pubsub emulator

```
proj=test
docker run -d -p 8085:8085 -e PUBSUB_PROJECT_ID=$proj adhawk/google-pubsub-emulator
unset GOOGLE_APPLICATION_CREDENTIALS
export PUBSUB_EMULATOR_HOST=localhost:8085

cd app
unique_key=test
go run . setup -p=$proj -s=$unique_key -t=$unique_key
go run . publish -p=$proj -s=$unique_key -t=$unique_key --batch=5 --times=4

```
