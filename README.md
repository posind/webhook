# Deplong Golang CI/CD to Google Cloud Platform

This is a simple Golang Model-Controller template using [Functions Framework for Go](https://github.com/GoogleCloudPlatform/functions-framework-go) and mongodb.com as the database host. It is compatible with Google Cloud Function CI/CD deployment.

Start here: Just [Fork this repo](https://github.com/gocroot/gcp/)

## MongoDB Preparation

The first thing to do is prepare a Mongo database using this template:

1. Sign up for mongodb.com and create one instance of Data Services of mongodb.
2. Go to Network Access menu > + ADD IP ADDRESS > ALLOW ACCESS FROM ANYWHERE  
   ![image](https://github.com/gocroot/gcp/assets/11188109/a16c5a73-ccdc-4425-8333-73c6fbf78e6d)  
3. Download [MongoDB Compass](https://www.mongodb.com/try/download/compass), connect with your mongo string URI from mongodb.com
4. Create database name iteung and collection reply  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/23ccddb7-bf42-42e2-baac-3d69f3a919f8)  
5. Import [this json](https://whatsauth.my.id/webhook/iteung.reply.json) into reply collection.  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/7a807d96-430f-4421-95fe-1c6a528ba428)  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/fd785700-7347-4f4b-b3b9-34816fc7bc53)  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/ef236b4d-f8f9-42c6-91ff-f6a7d83be4fc)  
6. Create a profile collection, and insert this JSON document with your 30-day token and WhatsApp number.  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/5b7144c3-3cdb-472b-8ab3-41fe86dad9cb)  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/829ae88a-be59-46f2-bddc-93482d0a4999)  

   ```json
   {
     "token":"v4.public.asoiduasoijfiun98erjg98egjpoikr",
     "phonenumber":"6281111222333"
   }
   ```

   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/06330754-9167-4bf4-a214-5d75dab7c60a)  

## Folder Structure

This boilerplate has several folders with different functions, such as:

* .github: GitHub Action yml configuration.
* config: all apps configuration like database, API, token.
* controller: all of the endpoints functions
* model: all of the type structs used in this app
* helper: helper folder with a list of functions only called by others file
* route: all routes URL

## GCP Cloud Function CI/CD setup

To get an auth in Google Cloud, you can do the following:

1. Open Cloud Shell Terminal, type this command line per line:  
   
   ![image](https://github.com/gocroot/gcp/assets/11188109/14f8e9d7-f74c-4f74-ab9c-72731a3e5f13)  

   ```sh
   # Get a List of Project IDs in Your GCP Account
   gcloud projects list --format="value(projectId)"
   # Set Project ID Variable
   PROJECT_ID=yourprojectid
   # Enable several api service
   gcloud services enable cloudfunctions.googleapis.com --project="${PROJECT_ID}"
   gcloud services enable cloudbuild.googleapis.com --project="${PROJECT_ID}"
   gcloud services enable run.googleapis.com --project="${PROJECT_ID}"

   # Create a service account
   gcloud iam service-accounts create "posind" --project "${PROJECT_ID}"
   # Create JSON key for GOOGLE_CREDENTIALS variable in GitHub repo
   gcloud iam service-accounts keys create "key.json" --iam-account "posind@${PROJECT_ID}.iam.gserviceaccount.com"
   # Read the key JSON file and copy the output, including the curl bracket, go to step 5.
   cat key.json
   # Authorize service account as CF Dev
   gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=serviceAccount:posind@${PROJECT_ID}.iam.gserviceaccount.com --role=roles/cloudfunctions.developer

   # Authorize service account to act as admin in Cloud Run service
   gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=serviceAccount:posind@${PROJECT_ID}.iam.gserviceaccount.com --role=roles/run.admin
   # Authorize service account to delete artifact registry
   gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=serviceAccount:posind@${PROJECT_ID}.iam.gserviceaccount.com --role=roles/artifactregistry.admin
   # Authorize service account to deploy cloud function
   gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=serviceAccount:posind@${PROJECT_ID}.iam.gserviceaccount.com --role=roles/cloudfunctions.developer
   gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=serviceAccount:posind@${PROJECT_ID}.iam.gserviceaccount.com --role=roles/logging.viewer
   ```

3. Open Menu Cloud Build>settings, select the Service Account created by step 1, and enable Cloud Function Developer.  
   ![image](https://github.com/gocroot/gcp/assets/11188109/3ebc81b6-18b7-4d44-90b4-0abf67f82d66)  
   ![image](https://github.com/gocroot/gcp/assets/11188109/d2628542-99a6-44ce-ba78-798c249e0f22)  
5. Go to the GitHub repository; in the settings, menu>secrets>action, add GOOGLE_CREDENTIALS vars with the value from the key.json file.
6. Add other Vars into the secret>action menu:  

   ```sh
   MONGOSTRING=mongodb+srv://user:pass@gocroot.wedrfs.mongodb.net/
   WAQRKEYWORD=yourkeyword
   WEBHOOKURL=https://asia-southeast1-PROJECT_ID.cloudfunctions.net/gocroot/webhook/inbox
   WEBHOOKSECRET=yoursecret
   WAPHONENUMBER=62811111
   ```
7. Rename gcf.yml into main.yml to activate Github Action

## WhatsAuth Signup

1. Go to the [WhatsAuth signup page](https://wa.my.id/) and scan with your WhatsApp camera menu for login.
2. Input the webhook URL(<https://yourappname.alwaysdata.net/whatsauth/webhook>) and your secret from the WEBHOOKSECRET setting environment on Always Data.  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/e0b5cb9d-e9b3-4d04-bbd5-b03bd12293da)  
3. Follow [this instruction](https://whatsauth.my.id/docs/), at the end of the instruction, you will get 30 days token using [this request](https://wa.my.id/apidocs/#/signup/signUpNewUser)
4. Save the token into MongoDB, open iteung db, and insert this JSON document with your 30-day token and WhatsApp number.  
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/829ae88a-be59-46f2-bddc-93482d0a4999)  
   ```json
   {
     "token":"v4.public.asoiduasoijfiun98erjg98egjpoikr",
     "phonenumber":"6281111222333"
   }
   ```
   ![image](https://github.com/gocroot/alwaysdata/assets/11188109/06330754-9167-4bf4-a214-5d75dab7c60a)  

## Refresh Whatsapp API Token

To continue using the WhatsAuth service, we must obtain a new token every three weeks before it expires in 30 days.

1. Open Menu Cloud Scheduler. You can just search it like the screenshot.  
   ![image](https://github.com/gocroot/gcp/assets/11188109/58e3f419-123b-4a69-89d2-9a1d3adb1b76)  
2. Click Create Job Input every 29 days; next, choose Target type HTTP, input refresh token URL from cloud function, HTTP method Get.  
   ![image](https://github.com/gocroot/gcp/assets/11188109/a9ee6af9-f8b6-404c-8a60-4c3df63b534e)  
   ![image](https://github.com/gocroot/gcp/assets/11188109/9b7d3f80-b264-4690-8776-9a8158a5f29c)    
3. Completing create schedule

## Upgrade Apps

If you want to upgrade apps, please delete (go.mod) and (go.sum) files first, then type the command in your terminal or cmd :

```sh
go mod init gocroot
go mod tidy
```


## Managing Module
Steps are follow:
1. Create Package in mod folder
2. Edit modcaller.go file, call your mod package here
3. Register keyword in module collection in mongodb
