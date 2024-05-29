## README

### CLI Build

1. **Clone the Repository**
   ```sh
   git clone <repository-url>
   cd <repository-directory>
   ```

2. **Build the CLI**
   ```sh
   cd cli
   go build -o yb_infra main.go
   ```

### Infrastructure Creation Using Terragrunt

1. **Install Terragrunt**
   Follow the [Terragrunt installation guide](https://terragrunt.gruntwork.io/docs/getting-started/install/).

2. **Navigate to the Infrastructure Directory**
   ```sh
   cd yb_infra/infra/env/dev
   ```

3. **Configure Terragrunt**
   Ensure you have the necessary inputs `config.hcl` configuration file. Here is an example configuration:
   ```hcl
   inputs = {
    environment         = "dev"
    aws_region          = "us-west-2"
    table_name_prefix   = "InstanceSchedules"
   }
   ```

4. **Apply the Infrastructure Configuration**
   ```sh
   terragrunt run-all apply -input=false -lock=false -auto-approve --terragrunt-non-interactive
   ```

### Lambda Function Build

1. **Navigate to the Lambda Function Directory**
   ```sh
   cd lambda
   ```

2. **Install Dependencies**
   ```sh
   GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
   ```

3. **Package the Lambda Function**
   ```sh
   zip yb_infra.zip bootstrap
   ```

4. **Deploy the Lambda Function**
   Reapply the Infrastructure Configuration to upload latest lambda code.

### Sample CLI Usage

#### Create Schedule

```sh
./yb_infra create-schedule --cloud-type aws --instance-id <instance-id> --start-time "08:00" --stop-time "20:00" --timezone "IST" --aws-region "us-west-2" --friendly-name "MyInstance"
```

#### List Schedules

```sh
./yb_infra list-schedules --cloud-type aws
```

#### Delete Schedule

```sh
./yb_infra delete-schedule --cloud-type aws --instance-id <instance-id>
```

### Environment Variables

Make sure to set the following environment variables before running the CLI:

```sh
export TABLE_NAME=your-dynamodb-table
```