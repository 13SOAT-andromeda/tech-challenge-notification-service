FROM public.ecr.aws/lambda/nodejs:20
COPY index.js .
CMD [ "index.handler" ]
