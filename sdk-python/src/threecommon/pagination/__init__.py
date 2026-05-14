"""Pagination module — exports the sync and async auto-paginating iterators.

User usually receive these from resource methods (e.g.
``client.events.list_auto_paginate(...)``) rather than constructing them
directly. Importing the submodule directly is supported::

    from threecommon.pagination import Iter, AsyncIter
"""

from threecommon.pagination.auto_paginator import AsyncIter, Iter

__all__ = ("AsyncIter", "Iter")
