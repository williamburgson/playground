from pyiceberg.schema import Schema
from pyiceberg.types import NestedField, StringType, FloatType, IntegerType
from pyiceberg.partitioning import PartitionField, PartitionSpec
from pyiceberg.transforms import BucketTransform, IdentityTransform
from pyiceberg.table.sorting import SortOrder, SortField

# data: ~/Documents/NYU/CUSP/PUI/World\ firearms\ murders\ and\ ownership.csv

schema = Schema(
    NestedField(field_id=1, name="Country/Territory", field_type=StringType()),
    NestedField(field_id=2, name="ISO code", field_type=StringType()),
    NestedField(
        field_id=3, name="Percent of homicides by firearm", field_type=FloatType()
    ),
    NestedField(
        field_id=4, name="Number of homicides by firearm", field_type=IntegerType()
    ),
    NestedField(
        field_id=5,
        name="Homicide by firearm rate per 100,000 pop",
        field_type=FloatType(),
    ),
    NestedField(field_id=6, name="Rank by rate of ownership", field_type=IntegerType()),
    NestedField(
        field_id=7, name="Average firearms per 100 people", field_type=FloatType()
    ),
    NestedField(
        field_id=8, name="Average total all civilian firearms", field_type=FloatType()
    ),
)

partition_spec = PartitionSpec(
    PartitionField(
        source_id=4,
        field_id=1000,
        transform=BucketTransform(10),
        name="num_homicides_bkt",
    )
)

sort_order = SortOrder(SortField(source_id=2, transform=IdentityTransform()))
