# Async API Request Builder Service

## Overview

This service is designed to send multiple asynchronous REST API requests to a specified server. It takes a folder name
as an argument, which contains a `data.json` file. This file defines the settings for the server, base URL, and
resources to request. The service manages a pool of simultaneous connections to prevent server overload and dynamically
adjusts the connection pool in case of errors. The results of the API requests are saved as a `.json` file in the same
directory.

---

## Features

1. **Multiple Asynchronous Requests**:
    - The service builds and sends multiple asynchronous API requests based on the provided configuration.
    - The maximum number of simultaneous connections is limited to avoid overwhelming the server.

2. **Error Handling**:
    - If a specified error occurs (e.g., HTTP 400), the number of concurrent threads is reduced by half to manage load
      and retry the requests.

3. **Result Output**:
    - The result of the requests is saved as a `.json` file in the same directory as the input `data.json` file.
    - Each result corresponds to the index of the original request, ensuring traceability.

---

## Usage

1. **Input Configuration**:
    - The service takes a folder name as an argument, which should contain the `data.json` configuration file.

2. **Structure of `data.json`**:
   The `data.json` file must include the following fields:

   ```json
   {
     "base_url": "demo.mediascout.ru",
     "url": "/webapi/clients/getclients",
     "ssl": "true",
     "login": "",
     "password": "",
     "ord": "ms",
     "connPool" : 50,
     "errlist": [
       "400"
     ],
     "headers": {
       "Content-type": "application/json"
     },
     "method": "post",
     "data": [
       {
         "Inn": "7717654289",
         "Status": "Active"
       },
       {
         "Inn": "7804347179",
         "Status": "Active"
       }
     ]
   }
   ```

- **base_url**: The base server URL to connect to.
- **url**: The endpoint to request.
- **ssl**: Set to `"true"` for HTTPS, otherwise use HTTP.
- **login/password**: Optional fields for authentication.
- **connPool**: Maximum number of simultaneous connections.
- **errlist**: A list of error codes (e.g., `"400"`) that will trigger a reduction in threads.
- **headers**: HTTP headers to be included in the request.
- **method**: HTTP method (e.g., `post`, `get`).
- **data**: The request payload, an array of JSON objects.

3. **Running the Service**:
   Run the service with the folder name containing the `data.json` file:
   ```bash
   ./asyncApi.go <folder_name>
   ```

   For example, to test with a folder named `adb808b6-1005-416e-a627-5bffebf074bc`, you would create a file `data.json`
   in that directory with the following content:

   ```json
   {
     "base_url": "demo.mediascout.ru",
     "url": "/webapi/clients/getclients",
     "ssl": "true",
     "login": "",
     "password": "",
     "ord": "ms",
     "connPool" : 50,
     "errlist": [
       "400"
     ],
     "headers": {
       "Content-type": "application/json"
     },
     "method": "post",
     "data": [
       {
         "Inn": "7717654289",
         "Status": "Active"
       },
       {
         "Inn": "7804347179",
         "Status": "Active"
       }
     ]
   }
   ```

---

## Result Format

The result of the API requests is saved as a JSON file in the same directory as the `data.json` file. The response is an
array of data objects, where each result corresponds to the index of the original request.

**Example Result Format**:

```json
{
  "data": [
    {
      "ActionType": "Contracting",
      "Amount": null,
      "Cid": null,
      "ClientId": "CL0IDLpNJkkEiozIrcjHAz4g",
      "ContractorId": "AG-vlmLcjPYEOhj1JTcIXusQ",
      "ContractorInn": "7709857180",
      "ContractorName": "ООО «Дэфт»",
      "index": 0
    }
  ]
}
```

---

## Logging

- All errors and critical information are logged in a specified log file.

---

## Testing

To test the service, create a folder (e.g., `adb808b6-1005-416e-a627-5bffebf074bc`) and place the following `data.json`
file in it:

```json
{
  "base_url": "demo.mediascout.ru",
  "url": "/webapi/clients/getclients",
  "ssl": "true",
  "login": "",
  "password": "",
  "ord": "ms",
  "connPool": 50,
  "errlist": [
    "400"
  ],
  "headers": {
    "Content-type": "application/json"
  },
  "method": "post",
  "data": [
    {
      "Inn": "7717654289",
      "Status": "Active"
    },
    {
      "Inn": "7804347179",
      "Status": "Active"
    }
  ]
}
```

Run the service, specifying the folder name:

```bash
./api-service adb808b6-1005-416e-a627-5bffebf074bc
```

The result will be saved as a `.json` file in the same directory.

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.