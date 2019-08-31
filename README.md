# How to run
docker-compose.yml:
```
version: "2.0"
volumes:
  uploads:
services:
  latex:
    image: aido93/latex-server:latest
    container_name: latex-server
    ports:
      - 8082:8080
    volumes:
      - uploads:/data
```

After that just
```
docker-compose up -d
```
# How to check
```
curl -X POST localhost:8082/v1/compile -H "Content-Type: multipart/form-data" -F "upload[]=@./tests/main.tex" -F "token=asdf" > asdf.pdf
```
Also it supports asyncronous mode. For using it you need do the following:
1. Connect services in your `docker-compose.yml`.
Example:
```
  latex:
    image: aido93/latex-server:latest
    container_name: latex-server
    volumes:
      - pdf_uploads:/data
    environment:
      # This line says to latex-server to use asyncronous mode.
      # So responses will be forwarded there
      CALLBACK_URL: http://app:8000
      DEBUG: "true"
  app:
    build: .
    container_name: app
    environment:
      # This line says to your app where is latex-server
      PDF_COMPILER_URL: http://latex-server:8080/v1/compile
```
2. Implement in your code sending request
```
import os
import string
import random
import requests
from jinja2 import Template
pdf_compiler_url=os.getenv("PDF_COMPILER_URL")
letters = string.ascii_lowercase
token=''.join(random.choice(letters) for i in range(64))
with open('templates/main.tex') as f:
    template = Template(f.read())
source=template.render(name='John')
# upload-pdf/ - is your view uri. The place in your code that works with uploading files
# You can use more than one uri for receiving answer (for instance, for diffrent situations).
# files MUST contain "upload[]" field that MUST contain at least 'main.tex' file
p=requests.post(pdf_compiler_url, files={"upload[]": ('main.tex', source.encode())}, data={"token": token, "uri": "upload-pdf/"})
print(p.content)
```
3. Implement in your code receiving answer
```
from django.http import JsonResponse
from django import forms
from .models import Report
from django.views.decorators.csrf import csrf_exempt

class UploadFileForm(forms.Form):
    token = forms.CharField(max_length=64)
    binary = forms.FileField()

@csrf_exempt
def upload_pdf(request):
    if request.method == 'POST':
        form = UploadFileForm(request.POST, request.FILES)
        if form.is_valid():
            token = form.cleaned_data['token']
            us=Report.objects.all().filter(token=token)
            if us.count()>1:
                return JsonResponse({"status": "found more than one report"})
            for u in us:
                u.pdf=request.FILES['binary']
                u.save()
                return JsonResponse({"status": "ok"})
            return JsonResponse({"status": "token not found"})
        return JsonResponse({"status": "invalid form"})
    return JsonResponse({"status": "undefined method"})

# urls.py:
from django.urls import include, path
from django.conf.urls import url

from . import views

urlpatterns = [
    path('upload-pdf/', views.upload_pdf, name='upload_pdf'),
]
```
