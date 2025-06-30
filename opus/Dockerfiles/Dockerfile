# --------------------------------------------------------------------------------------------------
# Builder Image
# --------------------------------------------------------------------------------------------------
# See ./builder for this image
FROM localhost/opus-builder AS builder

ARG VERSION
RUN set -eux \
	&& if [ "${VERSION}" = "latest" ]; then \
		pip3 install --no-cache-dir --no-compile ansible; \
	elif [ "${VERSION}" = "2.13" ]; then \
		pip3 install --no-cache-dir "ansible>=${VERSION},<2.14"; \
	elif [ "${VERSION}" = "2.14" ]; then \
		pip3 install --no-cache-dir "ansible>=${VERSION},<2.15"; \
	elif [ "${VERSION}" = "2.15" ]; then \
		pip3 install --no-cache-dir "ansible-core>=${VERSION},<2.16"; \
	else \
		pip3 install --no-cache-dir "ansible>=${VERSION},<$(echo "${VERSION}+0.1" | bc)"; \
	fi \
	\
	&& if [ "${VERSION}" != "latest" ]; then \
		ansible --version | grep ^ansible | grep -E "${VERSION}\.[0-9]+" \
		&& ansible-playbook --version | grep ^ansible | grep -E "${VERSION}\.[0-9]+" \
		&& ansible-galaxy --version | grep ^ansible | grep -E "${VERSION}\.[0-9]+"; \
	fi \
	\
	&& find /usr/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib64/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib64/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /local/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /local/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf

# Python packages (copied to final image)
RUN set -eux \
	&& pip3 install --no-cache-dir --no-compile \
		junit-xml \
		lxml \
		paramiko \
		docker \
		jsondiff \
		netaddr \
		Jinja2 \
		MarkupSafe \
		PyNaCl \
	    bcrypt \
		rich \
		ipmitools \
	&& find /usr/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib64/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib64/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /local/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /local/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf


# --------------------------------------------------------------------------------------------------
# Final Image
# --------------------------------------------------------------------------------------------------
FROM rockylinux/rockylinux:9 AS production
ARG VERSION
# https://github.com/opencontainers/image-spec/blob/master/annotations.md
#LABEL "org.opencontainers.image.created"=""
#LABEL "org.opencontainers.image.version"=""
#LABEL "org.opencontainers.image.revision"=""
LABEL "maintainer"="cjcshadowsan <christopher.james.coates@gmail.com>"
LABEL "org.opencontainers.image.authors"="cjcshadowsan <christopher.james.coates@gmail.com>"
LABEL "org.opencontainers.image.vendor"="cjcshadowsan"
LABEL "org.opencontainers.image.licenses"="MIT"
LABEL "org.opencontainers.image.url"="https://github.com/cjcshadowsan/opus-seed"
LABEL "org.opencontainers.image.documentation"="https://github.com/cjcshadowsan/opus-seed"
LABEL "org.opencontainers.image.source"="https://github.com/cjcshadowsan/opus-seed"
LABEL "org.opencontainers.image.ref.name"="Ansible ${VERSION} base"
LABEL "org.opencontainers.image.title"="Ansible ${VERSION} base"
LABEL "org.opencontainers.image.description"="Ansible ${VERSION} base"

RUN set -eux \
	&& dnf update -y \
	&& dnf install -y \
# libgcc required for ansible-vault
		libgcc \
		python3 \
		python3-pip \
# Additional tools for Ansible operations
		git \
		openssh-clients \
		sshpass \
		rsync \
		curl-minimal \
		wget \
	&& ln -sf /usr/bin/python3 /usr/bin/python \
	&& find /usr/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib64/ -name '__pycache__' -print0 | xargs -0 -n1 rm -rf \
	&& find /usr/local/lib64/ -name '*.pyc' -print0 | xargs -0 -n1 rm -rf

# Copy Python 3.12 from builder
COPY --from=builder /usr/local/bin/python3.12 /usr/local/bin/python3.12
COPY --from=builder /usr/local/bin/python3 /usr/local/bin/python3
COPY --from=builder /usr/local/bin/python /usr/local/bin/python
COPY --from=builder /usr/local/bin/pip3.12 /usr/local/bin/pip3.12
COPY --from=builder /usr/local/bin/pip3 /usr/local/bin/pip3
COPY --from=builder /usr/local/bin/pip /usr/local/bin/pip

# Copy Python site-packages from builder
COPY --from=builder /usr/local/lib/python3.12/ /usr/local/lib/python3.12/
# COPY --from=builder /usr/local/lib64/python3.12/site-packages/ /usr/local/lib64/python3.12/site-packages/
# COPY --from=builder /lib/python3/site-packages/ /lib/python3/site-packages/
# COPY --from=builder /lib64/python3/site-packages/ /lib64/python3/site-packages/

# Copy Ansible binaries from builder
COPY --from=builder /usr/local/bin/ansible /usr/local/bin/ansible
COPY --from=builder /usr/local/bin/ansible-config /usr/local/bin/ansible-config
COPY --from=builder /usr/local/bin/ansible-console /usr/local/bin/ansible-console
COPY --from=builder /usr/local/bin/ansible-doc /usr/local/bin/ansible-doc
COPY --from=builder /usr/local/bin/ansible-galaxy /usr/local/bin/ansible-galaxy
COPY --from=builder /usr/local/bin/ansible-inventory /usr/local/bin/ansible-inventory
COPY --from=builder /usr/local/bin/ansible-playbook /usr/local/bin/ansible-playbook
COPY --from=builder /usr/local/bin/ansible-pull /usr/local/bin/ansible-pull
COPY --from=builder /usr/local/bin/ansible-test /usr/local/bin/ansible-test
COPY --from=builder /usr/local/bin/ansible-vault /usr/local/bin/ansible-vault

# Update symlinks to use Python 3.12
RUN set -eux \
	&& ln -sf /usr/local/bin/python3.12 /usr/bin/python3 \
	&& ln -sf /usr/local/bin/python3.12 /usr/bin/python \
	&& ln -sf /usr/local/bin/pip3.12 /usr/bin/pip3 \
	&& ln -sf /usr/local/bin/pip3.12 /usr/bin/pip

# Pre-compile Python for better performance
RUN set -eux \
	&& python3 -m compileall /usr/lib/python3 || true

RUN set -eux \
    && python3 -m compileall /usr/lib64/python3 || true

RUN set -eux \
    && python3 -m compileall /usr/local/lib/python3.12 || true

RUN set -eux \
    && python3 -m compileall /usr/local/lib64/python3.12 || true

WORKDIR /data
CMD ["/bin/sh"]
