# Bukka Backend

Welcome to the Bukka Backend project! This repository contains the backend services for the Bukka application.

## Table of Contents

- [Introduction](#introduction)
- [Technologies Used](#technologies-used)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
  - [Authentication](#authentication)
    - [Register](#register)
    - [Login](#login)
    - [Profile](#profile)
    - [Rotate Token](#rotate-token)
    - [Verify Email](#verify-email)
    - [Resend Verification Email](#resend-verification-email)
    - [Delete Account](#delete-account)
    - [Request Account Recovery](#request-account-recovery)
    - [Recover Account](#recover-account)
    - [Change Password](#change-password)
    - [Forgot Password](#forgot-password)
    - [Reset Password](#reset-password)
    - [Home](#home)
- [Contributing](#contributing)
- [License](#license)

## Introduction

Bukka Backend is designed to provide robust and secure authentication services for the Bukka application. This documentation will guide you through the installation and usage of the current authentication endpoints.

## Technologies Used

- **Programming Language:** Go (Golang)
- **Framework:** Gin
- **Database:** PostgreSQL
- **Authentication:** Paseto
- **Other Tools:** Docker, Redis, Kubernetes

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/frankoe-dev/bukka_backend.git
    cd bukka_backend
    ```

2. Install dependencies:
    ```bash
    go mod tidy
    ```

3. Set up environment variables:
    Create a `.env` file in the root directory and add the following:
    ```env
    DATABASE_URL=your_database_url
    SECRET_KEY=your_secret_key
    ```

4. Start the necessary services:
    ```bash
    make start_ps
    make start_redis
    ```

5. Run database migrations:
    ```bash
    make migrateup
    ```

6. Start the server:
    ```bash
    go run main.go
    ```

## Usage

To use the authentication endpoints, you can use tools like [Postman](https://www.postman.com/) or [cURL](https://curl.se/). Below are the details of the available endpoints.

## API Endpoints

### Authentication

#### Home

- **Endpoint:** `/v1/auth/home`
- **Method:** `GET`
- **Description:** Basic endpoint to check the status of the server.
- **Response:**
    ```json
    {
        "msg": "Welcome to Bukka homepage 😃 Auth service. Your account is safe with us"
    }
    ```

#### Register

- **Endpoint:** `/v1/auth/register`
- **Method:** `POST`
- **Description:** Create a new user account.
- **Request Body:**
    ```json
    {
        "full_name": "Tim1",
        "email": "tomy8@gmail.com",
        "username": "timm8",
        "phone": "0912345678",
        "password": "12b3n12B$"
    }
    ```
- **Response:**
    ```json
    {
        "uid": "a3228666-d9e3-41b4-8438-ae47f4877189",
        "is_email_verified": false,
        "username": "timm8",
        "email": "tomy8@gmail.com",
        "created_at": "2024-06-19T17:55:26.793316Z",
        "password_changed_at": "0001-01-01T00:00:00Z",
        "is_suspended": false,
        "is_mfa_enabled": false,
        "is_deleted": false,
        "image_url": "",
        "access_token": "v2.local.Blbx3kh9J9pFRM_9mxHvvJtl20plIYO8nqT6HLmXzaU5djN4oWrzOsDXBqg4ADGwsb2RG9DYXEmevFgkjlN8pr7tOrVT7Ax32xy2m0hXDQ1LrNsPmsFT2dlAZ3LBbCyCoupArzNJtRW2MHxujuBBsVGrJ-BR2C9-u3rkAxO37zTgn2RmBi-V2yDeeFG37RKvERZdQskn48X9Qmf_YinmacINWLdkEeeW9fVWSH0SL8ZwgBcf0w2LXdno7xSSaW4T34CW11j3AzLkE5DIhj1E1U-DhjSVBDrDeHCwddJGctbwT9IU_gCcV30ajzJmS_AQgJA9kUseqixRAtB5VToRdcfv-YXpdYni5QJOo7Ig93gu33DUyx-s2HEfhmhbmKC_NOXZvwDkQBP5ykF3pUYDL4VogBmS-KKXGTUpCcXOZUje_tMn4_nvpJ1OdeekRpt8w8xZ28UecNrjorMCZetq1jADtNkHDMb3w9I1_CKJIR12FBDvH77UrD7Nl7v3GBLt_Mc97ppkK6AN23m1-h-V39AxInYzU7-nRqtJyKt3NhD0pJy_laJabd2bmhbQEK9gM7SZBfUG2uTQdr-RYRIodpFk5ATljxPrfedDSsHuQmVIWe3W8ck2yQWTXF5rjMfl7Tk.bnVsbA",
        "access_token_expires_at": "2024-06-19T19:55:26.935308+01:00",
        "refresh_token": "",
        "refresh_token_expires_at": "0001-01-01T00:00:00Z"
    }
    ```

#### Login

- **Endpoint:** `/v1/auth/login`
- **Method:** `POST`
- **Description:** Authenticate a user and return a JWT token.
- **Request Body:**
    ```json
    {
        "identifier": "tomy8@gmail.com",
        "password": "12b3n12B$"
    }
    ```
- **Response:**
    ```json
    {
        "uid": "a3228666-d9e3-41b4-8438-ae47f4877189",
        "is_email_verified": false,
        "username": "timm8",
        "email": "tomy8@gmail.com",
        "created_at": "2024-06-19T17:55:26.793316Z",
        "password_changed_at": "0001-01-01T00:00:00Z",
        "is_suspended": false,
        "is_mfa_enabled": false,
        "is_deleted": false,
        "image_url": "",
        "access_token": "v2.local.Ox0_MrmKTHLEcaVidmTDbeP2BAf4qUxkilksvBA_iAs4PF-IGPgtbbkHKGQWSqTCRpmrpg6x3YPgYdqJnhCi2STSd7YjAeFwcz8imV68IUZXbw_ZHkHWJmgl_BArgBIi4kyX4PBF_jeMp9HPFZWzjczFZcZBiF-pN9EDjsTRM6RJvyaCxCd5-P7m76xVzMgNGX2hflhLQTrxq-wkLu1Ok8s-Mn8ZSEBdKJixw9r0Hm0N7flLzzn_rxQWqSzM47wQNFuHspaMeSIJTuoQIkX2hl9PuaEk49bAuYS__HouWfLuTAIeCcGvx4DJDP3rsqFm2C7UTC0Jg16sD8cN2__pm5revYRATsKt7lyUaryKNjJ91w9C8rAJQ0cX-iwObH46g1IOB9Osc1MCFg96ZCoR7FKjlsO_pVwFkVLLwh6gIE-RfupQtBgkoiLI-JIgHOd_9H9XPZeajNUSyEq_jK8g-ZSrZhhrdzeDNTFAWk1iEsZRCmWRrje9cmo1qsuTxmNzWNhuv4gd4JeAhG9rNFFrUIWfdxTp3FeGr77H4c076_SFdwsfqYeKAuv6EByn5PivWuzPOhYdwoDYc1bzGE046YOiGZoRt2hheJpT8AzVNX7SR72nr8Dt56aqj0nW_1ZLxbs.bnVsbA",
        "access_token_expires_at": "2024-06-19T19:59:22.252449+01:00",
        "refresh_token": "v2.local.Co20hm1HvnfuOlqyZ3Z0oUK76bDbrWT_L1Uz5-9O1oMz0N768VOoOXNxehVPV3cse2Y3b8yDrt-2cDW_KYodJDCX7sgLdLRvoTzG51X74A7i5WkQYF0Yo1NTHdxKHolQ92nHFsESeGtwpQGTFQCCo5qXXbQQIicArnioKckw2BJTv7GKYhFAD0QGD89255Xs9FeTK8H4ipG33Y_4wUW6wioZ6rAH8hy4Oy02ImJ1MJQuxbRkHvzJjORiAkDEbsEj0LicGitOY_OPjQugUS28cxKsWB4D3zjzLoVKR6l6Fo9J3WXDTAaUPhf_6ZEs9pftuS0Mr9XmZzGmurp5I7vDAkHb-nYbz6R78RqjbmO-sgKuU_xbyvwN7j16LqIIk1HHJoLzDfVcCGwh-PNyrTOj7Owag_g_Dszd2wHSC2yfYt-lAjhqvX5OfRulez5XWinIG1eghJk0MKTB_A17iGLw8SyFVkOs2aDS_pXSIgZfV8lu4cRAq74jJWP7Xq4rn_Uzw1QYf5_5IAzpCxml2C6miw_WGxYSwEoz2A1dUpk7ZH2UHauGlFQFSIlUqwmWNgdO2h4ArvY7_uZHBUjt314HNjTekbwAyddiFKTmmA.bnVsbA",
        "refresh_token_expires_at": "2024-06-26T18:59:22.25059+01:00"
    }
    ```


#### Login with Google

- **Endpoint:** `/v1/auth/google/login?fcmToken=12345`
- **Method:** `GET`
- **Description:** Authenticate a user using their Google account and return a JWT token. The user will be redirected to the google screen to select their email. After that they are returned to /v1/auth/google/callback where the response data is returned

- **Response:**
    ```json
    {
        "uid":"c7d6c3c9-e730-431d-bd96-e4749416acc3","is_email_verified":true,
        "username":"",
        "email":"ekechukwuemeka25@gmail.com","created_at":"2024-07-18T13:55:53.138694Z","password_changed_at":"0001-01-01T00:00:00Z","is_suspended":false,
        "is_mfa_enabled":false,
        "is_deleted":false,
        "image_url":"",
        "access_token":"v2.local.TGWNS3bL_TylDk4-OYFDuA7bSO-XYRSdehXJ4qbjt7JmfTeOVhTk4aQoQGGujEPVBh9FRef3yQsiYy4Q6OipT1fuu_-VBKStBQnFeOJ9Z0dC4XAMqNcekswnqTYCXzyLV8LbpF5W3SJE_QxYGRofj_LDIfKJD07NqsWbunlCx2UjbQQaqZuaBBiszz621haWsZwGunubVIYpJHh99MeEuC3IVqxom2SO_FtBoZyQdhyw9RY8vg7SQdVcmQOq_bazP8l7XjVSy5k4nYxVYNMjJoqZXJNOob831iUS8nFAyWnwRNIP_ZbKZWgcrakxThpPihbH3iP0Id_71JCCJDD4pnvAnSulOmVW9lvtBKOwdHNuhCaFcsQUZFmKOb15h6tBeOZZIyFr9YQ6qKBQWIV1J88YOz7UlGcN57Ej7zKTMupGOuh_ZyiS81WQiQfqjOtiBtbTE-P8KhlpVoWw0BlbTbCyG3d807X7eOScNiqhxJlZoprw0eyFEmGsqv8pe3qoEsKg8NJzb5DoRyfYTGPZcI-Er-LbJvLKXAScIRRe0_D8ZqlFVAoHn_LC3yad3HNGT6ssg14MyDnULBJ8QcpXMBC1YVEW8rG6xKyf0-Zt_KyZk8s3EbHGKSfDJucaGg1CsipgmgFfKibZRZ1tKz-txt_lnd1JQyndYM1EaXfk1rcdgh82pYXwJOFnENGfbDSgjFA5oxRlbqgBD_AQ3gzUbiy-vpugxPvio80LBTEg5UT7jUM6Z1CtJ2I7uJvK0Q82gyhFTGA24e5EAbpfY9gh3UVmz8e4ogALHfrlCmVAxt1IRDxB7VIq2XHGqtVpXnMplcHK0ZBstEh_TISoRfpP62spM97P_X-zEvF1.bnVsbA","refresh_token":"v2.local.CkKtnzuOCsRqu4NLqZ4lTu2KqC2vow9GKFQ0Ps1V4PokLeFFqq2vfeb-WnPqAcXYsqebIErCw-hczPppWTPV44qAP1BvvkZHi_rkz3kCglntvuf5xkAQCYBO4buVQs6bG2c6lDCkLpICc6rj4Vlx5Bb-Yfbvk3RfQOMmKIGQbNEX0YPE10EZqNyHV9RfOmP9kwJ1KCgnVbj87ha6p-6-PyweN-pyzGGBYpMYStNXBse7CPH-6bz5S5AkFbBU4yhdBNKwFaxAwyAAh5XYLwK7ETCG5UUD_8lz23waUKAi1kZpYhyTH9VfYuJdSrxlp00CYAv0I29LDtFRInkNN7wOelDb838i_e8VggbLXgyST1HEbbrLLTScVSsynsgUtPTWsRiLZSaA6cRc-DLAwUYKvlpDTGiNPvxLMtr0kGzyTIWS76JMPhYvQM2rITUTvEun6SI8EhSgi77Ri-4ONpy7EEITaWdAQCyz9vPiZTBA4-gI976Zf0oFYDEAjqot5c37rwmrlweh2rcypMtb70JFnPT_UumUe-KBO95XLiQ3PVny3Hptgq3yeMplm1GNKkSiN9SGl2dLHub9AzU76nGBoOkG1CKTZ9puRptOcezqs1YH5kHMXXrtKQMeT3tMo5ykXnkgBriEfDTA8oPfGApN5FV8MGP2K4QTyLmlllV8b4KZAyndm8pXArFBN94nYf2U8EUrVrJVibPaxZ_tKhHoosEQMr-5bxcD_WnALvvLLVMPHrRSNTX2-8mQBCWk_pHc8hPazc2FgfUsFX6oa1tSItOTd4PDB2zvG3vE1F7ugNBSoJgJcBdca0w.bnVsbA",
        "access_token_expires_at":"2024-07-18T15:56:20.864149+01:00","refresh_token_expires_at":"2024-07-25T14:56:20.864128+01:00"
    }
    ```

#### Profile

- **Endpoint:** `/v1/auth/profile`
- **Method:** `GET`
- **Description:** Get the profile of the authenticated user.
- **Headers:**
    ```http
    Authorization: Bearer v2.local.Ox0_MrmKTHLEcaVidmTDbeP2BAf4qUxkilksvBA_iAs4PF-IGPgtbbkHKGQWSqTCRpmrpg6x3YPgYdqJnhCi2STSd7YjAeFwcz8imV68IUZXbw_ZHkHWJmgl_BArgBIi4kyX4PBF_jeMp9HPFZWzjczFZcZBiF-pN9EDjsTRM6RJvyaCxCd5-P7m76xVzMgNGX2hflhLQTrxq-wkLu1Ok8s-Mn8ZSEBdKJixw9r0Hm0N7flLzzn_rxQWqSzM47wQNFuHspaMeSIJTuoQIkX2hl9PuaEk49bAuYS__HouWfLuTAIeCcGvx4DJDP3rsqFm2C7UTC0Jg16sD8cN2__pm5revYRATsKt7lyUaryKNjJ91w9C8rAJQ0cX-iwObH46g1IOB9Osc1MCFg96ZCoR7FKjlsO_pVwFkVLLwh6gIE-RfupQtBgkoiLI-JIgHOd_9H9XPZeajNUSyEq_jK8g-ZSrZhhrdzeDNTFAWk1iEsZRCmWRrje9cmo1qsuTxmNzWNhuv4gd4JeAhG9rNFFrUIWfdxTp3FeGr77H4c076_SFdwsfqYeKAuv6EByn5PivWuzPOhYdwoDYc1bzGE046YOiGZoRt2hheJpT8AzVNX7SR72nr8Dt56aqj0nW_1ZLxbs.bnVsbA
    ```
- **Response:**
    ```json
    {
        "user": {
            "id": "user_id",
            "username": "your_username",
            "email": "your_email"
        }
    }
    ```

#### Rotate Token

- **Endpoint:** `/v1/auth/rotate_token`
- **Method:** `POST`
- **Description:** Rotate the user's authentication token.
- **Request Body:**
    ```json
    {
        "refresh_token": "v2.local.r74ac0BsZUVZanWd6PGSo4uBza16FyP5aXr6aZiqTnfl3ndaLmC8v2zgOO4vg74S6QLigicP_H9Q3n_9t_A73M2f8NC_cOvuVyjb7wIiOxZKiunYJPMtsjTX51IBIyGrDTrHbXI_pdRCqZc6wDHa1AOreenbdeVOBTw4XLWNu4pXnTsFN3fMjRDBOFqDJxVlQNyb4kO2vytNCE7OEUO3OmYjLZiozpfkXRHWxDMyRj87tNvkS4PFzQKlu2p9FL5mQoXjL8L3t8lsSEaXUwlBfku2zCVUABI5IbSC-YkO0r5h6s01kmShDlqur8upq_ITMmok22IlppV8AH2U8lSHD5A13vQlGsbkDDQBrpMJ31IZrnqz5EgKcC8eoWwKNf5-RDDWMM4LYB2e2WCmWUXs5hGb0u5Gyify5FZxig-dz19aQzh5X_fE-xO15Z3q7O8Eny1sIquFKEEY4NJIGtvrrmDkEDjyXgB37fcdZIMHu6KJOsV363MfTMk7V4tSe5xNs4kB-OXEWggyKNf5VYRwAXAgQYERi_dps3Spm0PfcWAbUO5mFmCWVJSnioOf6iylXaf7fQ5PQP_pQXu1lDv45qlUXEwgXnpnnkvR9WeW4g.bnVsbA"
    }
    ```
- **Response:**
    ```json
    {
        "access_token": "v2.local.6WNoYrGQzMOHxCaID5UzJ1g4-1v723cas294CbdO0YRlRryz5iIRezfEj5ma9xTr5tMgSwgLQ8sdvraq2qlDKyLoRGENGYyrVWB5ZhT6Ki1_iBVemHyNjIHXCWTkj_Jw07YiBUzOn1FR46hdrucu7ADXnp6WsXR5Dr6gjL5LJpsHnrVXWVablNdnsHDE2ECWnD1hmy9rE6Tyt63UmeBx_AgERQeNTXTQORUyWznZdZFHI_f9vQGvjJGcyQPLq98KRK_5CpepkGQRI7U9ulEA_UpYna1OCvXob_ZmMgwhnZbAHmBZsOFSjmreskpiVft0EhgVlUj82z919IdVFxnSSwR3pGztBYnh7xvlrKUCeOegOKuv6JXMjy5P56NmZ6aJeesdukBsRRSyCuBvgsh55u3Kc9hrmq8nDddJC-gKATPWbZiGI53Q1IJhPtaeJMpUU6CUs06ww7YCNMxlsnpKUBByNz4Cl8u8KggXYHTWl0b-4JzEdBh7gIUmGKR5_5nGHB9lDWFO8BApvjMgQz-HzfhjGQ6k09l4jkKHg7WHjG-OMCI1FMDqTt4kDf-X4cI0lCz5SGWIORpUJGSxBql8RvBZ5uTqPZf-qKwCHBJvwQpl0t7-pmPH4mrZuLGs9WNoRIA.bnVsbA",
        "access_token_expires_at": "2024-06-19T20:04:18.716997+01:00",
        "refresh_token": "v2.local.kSZaFJIFy0-2hLIFp94EImKkgY6es2HMosIaiu9izvmZ69X4qzp9UsTXMNo9zk7n2_-_ioMBfNPrfH2KfJkbpiE4cj2JgxDTMlhIjO4Rnrar0cG5N34q3KhDbe79qN_JhSNJMZ7xF1GaAiOW1CRZ-4eulAarQ8LhHFyzvN7rh7cnQ57lUnksHjN-5Oh4e9sNDNg6dnVM4SC9Vro9uOQ8_Lr7uOcIrZKND5XKwR6_Ir4ByaBsVcQjpxa8vuWAqtufIydhQPrTIXOUHMjw1s-B5UniAo01mkIIVL32mlC_joxQG7iyE8Vz0uV8IqkmKP9YJZo21HnOrv5909QQ9HcyJDUO8-onqOwhgF5LY8-Y6_lt-B5uqeZRRRliK454QBFqIr5bdcweN4AcFB4NU542WzXBuUTlljA_zfEDKq1aTzLgEedC8EBG96g6xS2xqUExNmKokeMwQ4Hb5xvQFxlnewAapwNDZufFBpOOxFSuyzyx3gE77h27coQBGmfcPkSpptcEwQzPpHGEPhaLuu_XBhFox4ey7KAPqCgtH8VCFBFvpqG-yiKzBdx6PGavHkvJ0iKfURiRGTlbTb6-g2z_t0xW3mRPsvlTHEVy_zrxeg.bnVsbA",
        "refresh_token_expires_at": "2024-06-26T19:04:18.717006+01:00"
    }
    ```

#### Verify Email

- **Endpoint:** `/v1/auth/verify_email`
- **Method:** `POST`
- **Description:** Verify a user's email address.
- **Request Body:**
    ```json
    {
        "token": "634808"
    }
    ```
- **Response:**
    ```json
    {
        "msg": "Email verified successfully",
        "verified": true
    }
    ```

#### Resend Verification Email

- **Endpoint:** `/v1/auth/resend_verification`
- **Method:** `GET`
- **Description:** Resend the email verification link to the user.
- **Headers:**
    ```http
    Authorization: Bearer v2.local.6WNoYrGQzMOHxCaID5UzJ1g4-1v723cas294CbdO0YRlRryz5iIRezfEj5ma9xTr5tMgSwgLQ8sdvraq2qlDKyLoRGENGYyrVWB5ZhT6Ki1_iBVemHyNjIHXCWTkj_Jw07YiBUzOn1FR46hdrucu7ADXnp6WsXR5Dr6gjL5LJpsHnrVXWVablNdnsHDE2ECWnD1hmy9rE6Tyt63UmeBx_AgERQeNTXTQORUyWznZdZFHI_f9vQGvjJGcyQPLq98KRK_5CpepkGQRI7U9ulEA_UpYna1OCvXob_ZmMgwhnZbAHmBZsOFSjmreskpiVft0EhgVlUj82z919IdVFxnSSwR3pGztBYnh7xvlrKUCeOegOKuv6JXMjy5P56NmZ6aJeesdukBsRRSyCuBvgsh55u3Kc9hrmq8nDddJC-gKATPWbZiGI53Q1IJhPtaeJMpUU6CUs06ww7YCNMxlsnpKUBByNz4Cl8u8KggXYHTWl0b-4JzEdBh7gIUmGKR5_5nGHB9lDWFO8BApvjMgQz-HzfhjGQ6k09l4jkKHg7WHjG-OMCI1FMDqTt4kDf-X4cI0lCz5SGWIORpUJGSxBql8RvBZ5uTqPZf-qKwCHBJvwQpl0t7-pmPH4mrZuLGs9WNoRIA.bnVsbA
    ```
- **Response:**
    ```json
    {
        "msg": "Verification email sent",
    }
    ```

#### Delete Account

- **Endpoint:** `/v1/auth/delete_account`
- **Method:** `DELETE`
- **Description:** Delete a user's account.
- **Headers:**
    ```http
    Authorization: Bearer v2.local.6WNoYrGQzMOHxCaID5UzJ1g4-1v723cas294CbdO0YRlRryz5iIRezfEj5ma9xTr5tMgSwgLQ8sdvraq2qlDKyLoRGENGYyrVWB5ZhT6Ki1_iBVemHyNjIHXCWTkj_Jw07YiBUzOn1FR46hdrucu7ADXnp6WsXR5Dr6gjL5LJpsHnrVXWVablNdnsHDE2ECWnD1hmy9rE6Tyt63UmeBx_AgERQeNTXTQORUyWznZdZFHI_f9vQGvjJGcyQPLq98KRK_5CpepkGQRI7U9ulEA_UpYna1OCvXob_ZmMgwhnZbAHmBZsOFSjmreskpiVft0EhgVlUj82z919IdVFxnSSwR3pGztBYnh7xvlrKUCeOegOKuv6JXMjy5P56NmZ6aJeesdukBsRRSyCuBvgsh55u3Kc9hrmq8nDddJC-gKATPWbZiGI53Q1IJhPtaeJMpUU6CUs06ww7YCNMxlsnpKUBByNz4Cl8u8KggXYHTWl0b-4JzEdBh7gIUmGKR5_5nGHB9lDWFO8BApvjMgQz-HzfhjGQ6k09l4jkKHg7WHjG-OMCI1FMDqTt4kDf-X4cI0lCz5SGWIORpUJGSxBql8RvBZ5uTqPZf-qKwCHBJvwQpl0t7-pmPH4mrZuLGs9WNoRIA.bnVsbA
    ```
- **Response:**
    ```json
    {
        "msg": "Account deleted successfully"
    }
    ```

#### Request Account Recovery

- **Endpoint:** `/v1/auth/request_account_recovery`
- **Method:** `POST`
- **Description:** Request account recovery for a user.
- **Request Body:**
    ```json
    {
        "email": "tomy8@gmail.com"
    }
    ```
- **Response:**
    ```json
    {
        "msg": "URL has been sent to your email"
    }
    ```

#### Recover Account

- **Endpoint:** `/v1/auth/recover_account`
- **Method:** `GET`
- **Description:** Complete the account recovery process.
- **Request Query Parameters:**
    ```http
    token=Zz9R8leE-cikArbVH0eiV6VJL5tzYtC80bVlRpVmLTtOqJ7_04wL
    ```
- **Response:**
    ```json
    {
        "msg": "Account recovered, you can now login to access your account."
    }
    ```

#### Change Password

- **Endpoint:** `/v1/auth/change_password`
- **Method:** `POST`
- **Description:** Change the user's password.
- **Request Body:**
    ```json
    {
        "old_password": "12b3n12B$",
        "new_password": "12b3n12Bk$"
    }
    ```
- **Response:**
    ```json
    {
        "msg": "password changed successfully"
    }
    ```

#### Forgot Password

- **Endpoint:** `/v1/auth/forgot_password`
- **Method:** `POST`
- **Description:** Request a password reset link.
- **Request Body:**
    ```json
    {
        "email": "tomy7@gmail.com"
    }
    ```
- **Response:**
    ```json
    {
        "msg": "if an account exists a password reset email will be sent to you"
    }
    ```

#### Reset Password

- **Endpoint:** `/v1/auth/reset_password`
- **Method:** `POST`
- **Description:** Reset the user's password using the reset token.
- **Request Body:**
    ```json
    {
        "new_password": "12b3n12Bk$",
        "token": "785588"
    }
    ```
- **Response:**
    ```json
    {
        "msg": "Password changed successfully"
    }
    ```


## Contributing

We welcome contributions from the community. Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Create a new Pull Request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
