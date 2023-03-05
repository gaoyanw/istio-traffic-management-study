# 1. check deployment sar
kubectl create -f - -o yaml << EOF
apiVersion: authorization.k8s.io/v1
kind: SubjectAccessReview
metadata:
  creationTimestamp: null
spec:
  resourceAttributes:
    group: apps
    namespace: bookstoreserver
    resource: deployments
    verb: create
  user: system:serviceaccount:api-client:api-client-sa
status:
  allowed: true
  reason: 'RBAC: allowed by ClusterRoleBinding "deployment-admin-binding" of ClusterRole
    "deployment-admin" to ServiceAccount "api-client-sa/api-client"'
EOF

# 2.1. shelf admin can get
kubectl create -f - -o yaml << EOF

apiVersion: authorization.k8s.io/v1

kind: SubjectAccessReview

spec:

 user: system:serviceaccount:ns-foo:shelf-admin-sa
 resourceAttributes:

   group: gdch-bookstore.googleapis.com

   resource: shelves

   verb: get

   namespace: dev

EOF

# 2.2 shelf viewer can als create
kubectl create -f - -o yaml << EOF

apiVersion: authorization.k8s.io/v1

kind: SubjectAccessReview

spec:

 user: system:serviceaccount:ns-foo:shelf-admin-sa
 resourceAttributes:

   group: gdch-bookstore.googleapis.com

   resource: shelves

   verb: create

   namespace: dev

EOF

# 3.1. shelf viewer can get
kubectl create -f - -o yaml << EOF

apiVersion: authorization.k8s.io/v1

kind: SubjectAccessReview

spec:

 user: system:serviceaccount:ns-foo:shelf-viewer-sa
 resourceAttributes:

   group: gdch-bookstore.googleapis.com

   resource: shelves

   verb: get

   namespace: dev

EOF

# 3.2. shelf admi can NOT create
kubectl create -f - -o yaml << EOF

apiVersion: authorization.k8s.io/v1

kind: SubjectAccessReview

spec:

 user: system:serviceaccount:ns-foo:shelf-viewer-sa
 resourceAttributes:

   group: gdch-bookstore.googleapis.com

   resource: shelves

   verb: create

   namespace: dev

EOF