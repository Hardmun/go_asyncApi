```yaml

1) parameter for testing
   "adb808b6-1005-416e-a627-5bffebf074bc"

2) for test create json file data.json in catalog adb808b6-1005-416e-a627-5bffebf074bc

   {
   "base_url": "demo.mediascout.ru",
   "url": "/webapi/Invoices/GetInvoices",
   "ssl": "true",
   "login": "login",
   "password": "password",
   "headers": {
   "Content-type": "application/json"
   },
   "method": "post",
   "data": [
   {
   "Number": "УС31-08/695",
   "DateStart": "2023-08-31",
   "DateEnd": "2023-08-31"
   }
   ]
   }

3) compile excelFusion.py to exe file
   pyinstaller -F -w --icon=asyncApi.ico asyncApi.py


                                              ©DENTSU 2023
