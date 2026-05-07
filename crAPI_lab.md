**How to test the API**

Testing the tool in crAPI running in http://localhost:8888

1. First step login to crAPI and explore using BURP suite
2. second find the API Endpoints
3. Found the API endpoints.

IDOR endpoints
==============
1.GET /workshop/api/shop/orders/7
    Need any user Authentication Token to access the other people orders
Step:1 login as normal user 
step:2 go to shop and click on the past orders
step3: change the id in the url for other orders.
step4: may notice the other people order details

2.Authentication bugs
======================

